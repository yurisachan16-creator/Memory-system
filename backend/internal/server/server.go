package server

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"

	appconfig "github.com/yurisachan16-creator/Memory-system/backend/internal/config"
	"github.com/yurisachan16-creator/Memory-system/backend/internal/handler"
	"github.com/yurisachan16-creator/Memory-system/backend/internal/middleware"
	"github.com/yurisachan16-creator/Memory-system/backend/internal/repository"
)

type App struct {
	config appconfig.Config
	engine *gin.Engine
	db     *sql.DB
	redis  *redis.Client
}

func New(cfg appconfig.Config) (*App, error) {
	if strings.EqualFold(cfg.App.Env, "production") {
		gin.SetMode(gin.ReleaseMode)
	}

	db, err := repository.NewMySQL(cfg.Database)
	if err != nil {
		return nil, err
	}

	redisClient, err := repository.NewRedis(cfg.Redis)
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	if cfg.App.AutoMigrate {
		if err := repository.RunMigrations(cfg.Database, "file://migrations"); err != nil {
			_ = redisClient.Close()
			_ = db.Close()
			return nil, fmt.Errorf("run migrations: %w", err)
		}
	}

	engine := gin.New()
	engine.Use(
		middleware.RequestID(),
		middleware.RateLimit(),
		middleware.Logger(),
		middleware.Recovery(),
		middleware.CORS(),
	)
	handler.RegisterRoutes(engine, handler.Dependencies{
		DB:    db,
		Redis: redisClient,
	})

	return &App{
		config: cfg,
		engine: engine,
		db:     db,
		redis:  redisClient,
	}, nil
}

func (a *App) Run() error {
	return a.engine.Run(a.config.App.Address())
}

func (a *App) Close() {
	if a.redis != nil {
		_ = a.redis.Close()
	}

	if a.db != nil {
		_ = a.db.Close()
	}
}
