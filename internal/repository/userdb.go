package repository

import (
	"database/sql"

	"github.com/TemirB/rest-api-marketplace/internal/user"
	"go.uber.org/zap"
)

func (r *Repository) CreateUser(user *user.User) error {
	query := "INSERT INTO users (login, password) VALUES (?, ?)"
	_, err := r.DB.Exec(query, user.Login, user.Password)
	if err != nil {
		errorMsg := mapedError(err)
		r.logger.Error(
			errorMsg,
			zap.String("login", user.Login),
			zap.Error(err),
		)
		return err
	}

	return nil
}

func (db *Database) GetUser(Login string) (*user.User, error) {
	query := "SELECT login, password FROM users WHERE login = ?"
	row := db.DB.QueryRow(query, Login)

	var user user.User
	err := row.Scan(&user.Login, &user.Password)
	if err != nil {
		if err == sql.ErrConnDone {
			db.logger.Error(
				"Connection to the database is closed",
				zap.String("login", Login),
				zap.Error(err),
			)
			return nil, err
		} else {
			errorMsg := mapedError(err)
			db.logger.Warn(
				errorMsg,
				zap.String("login", Login),
				zap.Error(err),
			)
		}
		return nil, err
	}

	return &user, nil
}

func mapedError(err error) string {
	switch err {
	case sql.ErrNoRows:
		return "User not found"
	case sql.ErrTxDone:
		return "Transaction is already committed or rolled back"
	}
	return "An unexpected error occurred"
}
