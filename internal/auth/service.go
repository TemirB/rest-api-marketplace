package auth

import (
	"errors"

	"github.com/TemirB/rest-api-marketplace/pkg/hash"
	"github.com/TemirB/rest-api-marketplace/pkg/jwt"
)

var (
	InvalidLogin            = errors.New("login must be 3-50 characters")
	InvalidPassword         = errors.New("password must be at least 8 characters")
	UserAlreadyExists       = errors.New("user already exists")
	FailedToEncryptPassword = errors.New("failed to encrypt password")

	InvalidCredentials = errors.New("invalid credentials")
)

type Service struct {
	storage *storage
	manager *jwt.Manager
}

func NewService(
	storage *storage,
	tokenGenerator *jwt.Manager,
) *Service {
	return &Service{
		storage: storage,
		manager: tokenGenerator,
	}
}

func (s *Service) Register(login, password string) error {
	if validateLogin(login) {
		return InvalidLogin
	}

	if validatePassword(password) {
		return InvalidPassword
	}

	if s.validateUser(login) {
		return UserAlreadyExists
	}

	EncryptedPassword, err := hash.EncryptPassword(password)
	if err != nil {
		return FailedToEncryptPassword
	}

	user := User{
		Login:    login,
		Password: string(EncryptedPassword),
	}

	return s.storage.Create(&user)
}

func (s *Service) Login(login, password string) (string, error) {
	user, err := s.storage.GetByLogin(login)
	if err != nil {
		return "", InvalidCredentials
	}

	if hash.ComparePasswords(user.Password, password) {
		return "", InvalidCredentials
	}

	return s.manager.GenerateToken(login)
}

func (s *Service) ValidateToken(tokenString string) (string, error) {
	return s.manager.ValidateToken(tokenString)
}
