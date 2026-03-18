package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"gopkg.in/yaml.v3"
)

type composeContract struct {
	Services map[string]composeService `yaml:"services"`
}

type composeService struct {
	ContainerName string `yaml:"container_name"`
}

func TestDockerComposeUsesProjectScopedContainerNames(t *testing.T) {
	t.Parallel()

	composePath := repoComposePath(t)
	content, err := os.ReadFile(composePath)
	if err != nil {
		t.Fatalf("read docker compose file: %v", err)
	}

	var compose composeContract
	if err := yaml.Unmarshal(content, &compose); err != nil {
		t.Fatalf("parse docker compose file: %v", err)
	}

	for _, serviceName := range []string{"mysql", "redis", "backend", "frontend"} {
		service, ok := compose.Services[serviceName]
		if !ok {
			t.Fatalf("docker compose is missing %q service", serviceName)
		}
		if service.ContainerName != "" {
			t.Fatalf("service %q must not set container_name; rely on compose-managed project scoping instead", serviceName)
		}
	}
}

func repoComposePath(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve current file path")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "..", "docker-compose.yml"))
}
