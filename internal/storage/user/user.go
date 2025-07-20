package storage

import (
	"database/sql"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/TemirB/rest-api-marketplace/internal/pkg/db"
	"github.com/TemirB/rest-api-marketplace/internal/pkg/models/user"
)

type userStorage struct {
	repository *db.Repository
}

func NewStorage(repository *db.Repository) *userStorage {
	return &userStorage{
		repository: repository,
	}
}

func (r *userStorage) Create(user *user.User) error {
	query := `INSERT INTO users (login, password) VALUES ($1, $2)`
	_, err := r.repository.DB.Exec(query, user.Login, user.Password)
	if err != nil {
		r.repository.Logger.Error("Failed to create user", zap.Error(err))
		return errors.Errorf("failed to create user: %d", err)
	}
	return nil
}

func (r *userStorage) GetByLogin(login string) (*user.User, error) {
	query := `SELECT login, password FROM users WHERE login = $1`
	row := r.repository.DB.QueryRow(query, login)

	var user user.User
	err := row.Scan(&user.Login, &user.Password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.repository.Logger.Debug("User not found", zap.String("login", login))
			return nil, errors.New("user not found")
		}
		r.repository.Logger.Error("Failed to get user", zap.Error(err))
		return nil, errors.Errorf("failed to get user: %d", err)
	}

	return &user, nil
}

func (r *userStorage) Exists(login string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM users WHERE login=$1)"
	err := r.repository.DB.QueryRow(query, login).Scan(&exists)
	if err != nil {
		return false, errors.Errorf("failed to check if user exists: %d", err)
	}
	return exists, nil
}

func (r *userStorage) Delete(login string) error {
	query := `DELETE FROM users WHERE login = $1`
	_, err := r.repository.DB.Exec(query, login)
	if err != nil {
		r.repository.Logger.Error("Failed to delete user", zap.Error(err))
		return errors.Errorf("failed to delete user: %d", err)
	}
	return nil
}
