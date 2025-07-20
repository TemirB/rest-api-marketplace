package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/TemirB/rest-api-marketplace/config"
)

type Repository struct {
	DB     *sql.DB
	Logger *zap.Logger
}

func NewPostgresDB(cfg *config.Config, logger *zap.Logger) (*Repository, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open db connection: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	return &Repository{
		DB:     db,
		Logger: logger,
	}, nil
}

func (r *Repository) Connect() error {
	if r.DB == nil {
		return fmt.Errorf("db connection is not initialized")
	}

	if err := r.DB.Ping(); err != nil {
		return fmt.Errorf("failed to connect to db: %w", err)
	}
	return nil
}

func (r *Repository) Close() error {
	if r.DB != nil {
		if err := r.DB.Close(); err != nil {
			return fmt.Errorf("failed to close db connection: %w", err)
		}
	}
	return nil
}
