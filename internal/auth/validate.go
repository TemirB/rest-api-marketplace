package auth

import (
	"database/sql"
	"fmt"
)

// tested in service_test.go

func validPassword(password string) bool {
	// Нужно дописать логику валидации пароля
	// Например проверить, что пароль не содержит пробелов etc...
	if len(password) < 8 {
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

// Возвращает (true, nil), если пользователь найден;
// (false, nil) — если не найден;
// (false, err) — при ошибке работы с БД.
func (s *Service) userExists(login string) (bool, error) {
	exists, err := s.storage.Exists(login)

	if err == nil || err == sql.ErrNoRows {
		return exists, nil
	}
	return false, fmt.Errorf("checking user existence: %w", err)
}
