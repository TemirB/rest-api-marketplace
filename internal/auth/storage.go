package auth

// mockgen  -source=storage.go -destination=storage_mock_test.go -package=auth

import (
	"database/sql"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Repository interface {
	Exec(query string, args ...any) (sql.Result, error)
	QueryRow(query string, args ...any) *sql.Row
}

type Storage struct {
	repository Repository
	logger     *zap.Logger
}

func NewStorage(repo Repository, logger *zap.Logger) *Storage {
	return &Storage{
		repository: repo,
		logger:     logger,
	}
}

func (r *Storage) Create(user *User) error {
	query := `INSERT INTO users (login, password) VALUES ($1, $2)`
	_, err := r.repository.Exec(query, user.Login, user.Password)
	if err != nil {
		r.logger.Error("Failed to create user", zap.Error(err))
		return errors.Errorf("failed to create user: %d", err)
	}
	return nil
}

func (r *Storage) GetByLogin(login string) (*User, error) {
	query := `SELECT login, password FROM users WHERE login = $1`
	row := r.repository.QueryRow(query, login)

	var user User
	err := row.Scan(&user.Login, &user.Password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Debug("User not found", zap.String("login", login))
			return nil, errors.New("user not found")
		}
		r.logger.Error("Failed to get user", zap.Error(err))
		return nil, errors.Errorf("failed to get user: %d", err)
	}

	return &user, nil
}

func (r *Storage) Exists(login string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM users WHERE login=$1)"
	err := r.repository.QueryRow(query, login).Scan(&exists)
	if err != nil {
		return false, errors.Errorf("failed to check if user exists: %d", err)
	}
	return exists, nil
}

func (r *Storage) Delete(login string) error {
	query := `DELETE FROM users WHERE login = $1`
	_, err := r.repository.Exec(query, login)
	if err != nil {
		r.logger.Error("Failed to delete user", zap.Error(err))
		return errors.Errorf("failed to delete user: %d", err)
	}
	return nil
}
