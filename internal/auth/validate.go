package auth

import (
	"database/sql"
	"fmt"
	"regexp"
)

var (
	// Логин: 3–50 символов, только латиница, цифры, underscore, дефис,
	// обязательно начинается и заканчивается на букву или цифру
	loginRe = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]{1,48}[A-Za-z0-9]$`)

	// Пароль: минимум 8 символов, без пробелов,
	// обязательно хотя бы одна цифра, одна нижняя и одна верхняя буква, один спецсимвол
	digitRe   = regexp.MustCompile(`[0-9]`)
	lowerRe   = regexp.MustCompile(`[a-z]`)
	upperRe   = regexp.MustCompile(`[A-Z]`)
	specialRe = regexp.MustCompile(`[!@#\$%\^&\*\(\)\-_\+=\[\]{}|;:'",.<>\/?]`)
)

// tested in service_test.go

func validPassword(password string) bool {
	if len(password) < 8 || regexp.MustCompile(`\s`).MatchString(password) {
		return false
	}
	if !digitRe.MatchString(password) {
		return false
	}
	if !lowerRe.MatchString(password) {
		return false
	}
	if !upperRe.MatchString(password) {
		return false
	}
	if !specialRe.MatchString(password) {
		return false
	}
	return true
}

func validLogin(login string) bool {
	return loginRe.MatchString(login)
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
