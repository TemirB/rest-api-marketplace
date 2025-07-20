package auth

func validatePassword(password string) bool {
	// Нужно дописать логику валидации пароля
	if len(password) == 0 {
		return false
	}
	return true
}

func validateLogin(login string) bool {
	// Нужно дописать логику валидации логина
	if len(login) < 3 || len(login) > 50 {
		return false
	}
	return true
}

func (s *Service) validateUser(login string) bool {
	exists, _ := s.storage.Exists(login)
	return exists
}
