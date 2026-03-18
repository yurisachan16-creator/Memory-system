CREATE TABLE IF NOT EXISTS memories (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id VARCHAR(64) NOT NULL,
    content TEXT NOT NULL,
    category ENUM('preference', 'identity', 'goal', 'context') NOT NULL,
    source ENUM('chat', 'manual', 'system') NOT NULL,
    importance TINYINT NOT NULL DEFAULT 3 CHECK (importance BETWEEN 1 AND 5),
    content_hash VARCHAR(64) NOT NULL COMMENT '用于去重',
    is_deleted TINYINT(1) NOT NULL DEFAULT 0 COMMENT '软删除标记',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_category (user_id, category),
    INDEX idx_user_importance (user_id, importance DESC),
    INDEX idx_user_created (user_id, created_at DESC),
    INDEX idx_content_hash (user_id, content_hash),
    FULLTEXT INDEX ft_content (content)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

