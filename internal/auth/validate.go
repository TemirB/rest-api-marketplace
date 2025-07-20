package auth

import (
	"database/sql"
	"errors"
	"fmt"
)

func validPassword(password string) bool {
	// Нужно дописать логику валидации пароля
	// Например проверить, что пароль не содержит пробелов etc...
	if len(password) == 0 {
		return false
	}
	return true
}

func validLogin(login string) bool {
	// Нужно дописать логику валидации логина
	if len(login) < 3 || len(login) > 50 {
		return false
	}
	return true
}

func (s *Service) userExist(login string) (bool, error) {
	exists, err := s.storage.Exists(login)
	if err != nil {
		if errors.Is(err, sql.ErrConnDone) {
			return false, fmt.Errorf("database connection error: %w", err)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return exists, nil
}
