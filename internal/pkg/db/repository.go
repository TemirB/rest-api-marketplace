package repository

import (
	"database/sql"

	"go.uber.org/zap"
	"honnef.co/go/tools/config"
)

type Repository struct {
	DB     *sql.DB
	logger *zap.Logger
}

func NewDatabase(db *sql.DB, logger *zap.Logger) *Repository {
	return &Repository{
		DB:     db,
		logger: logger,
	}
}

func (d *Repository) Connect(cfg config.Config) error {
	err := config.Load
}
