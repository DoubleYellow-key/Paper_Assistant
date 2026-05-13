package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func New(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}

	db.SetMaxOpenConns(30)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping mysql: %w", err)
	}
	return db, nil
}

func Migrate(ctx context.Context, db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
			username VARCHAR(64) NOT NULL,
			email VARCHAR(128) NOT NULL UNIQUE,
			password_hash VARCHAR(255) NOT NULL,
			role VARCHAR(16) NOT NULL DEFAULT 'user',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
		`CREATE TABLE IF NOT EXISTS papers (
			id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
			user_id BIGINT UNSIGNED NOT NULL,
			title VARCHAR(512) NOT NULL,
			file_name VARCHAR(256) NOT NULL,
			file_path VARCHAR(512) NOT NULL,
			storage_path VARCHAR(512) NOT NULL,
			file_size BIGINT NOT NULL,
			parse_status VARCHAR(32) NOT NULL DEFAULT 'pending',
			parse_error TEXT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_papers_user_created (user_id, created_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
		`CREATE TABLE IF NOT EXISTS parse_jobs (
			id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
			paper_id BIGINT UNSIGNED NOT NULL UNIQUE,
			status VARCHAR(32) NOT NULL DEFAULT 'queued',
			progress TINYINT UNSIGNED NOT NULL DEFAULT 0,
			retry_count TINYINT UNSIGNED NOT NULL DEFAULT 0,
			max_retries TINYINT UNSIGNED NOT NULL DEFAULT 3,
			started_at DATETIME NULL,
			finished_at DATETIME NULL,
			error_message TEXT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
		`CREATE TABLE IF NOT EXISTS paper_parsed_texts (
			paper_id BIGINT UNSIGNED PRIMARY KEY,
			content LONGTEXT NOT NULL,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
	}

	for _, stmt := range stmts {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("migrate mysql failed: %w", err)
		}
	}
	return nil
}
