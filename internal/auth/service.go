package auth

import (
	"errors"

	"github.com/TemirB/rest-api-marketplace/internal/pkg/models/user"
	"github.com/TemirB/rest-api-marketplace/internal/pkg/secure"
	"github.com/TemirB/rest-api-marketplace/internal/pkg/token"
)

var (
	InvalidLogin            = errors.New("login must be 3-50 characters")
	InvalidPassword         = errors.New("password must be at least 8 characters")
	UserAlreadyExists       = errors.New("user already exists")
	FailedToEncryptPassword = errors.New("failed to encrypt password")

	InvalidCredentials = errors.New("invalid credentials")
)

type storage interface {
	Create(user *user.User) error
	GetByLogin(login string) (*user.User, error)
	Exists(login string) (bool, error)
	Delete(login string) error
}

type Service struct {
	storage   storage
	generator *token.Generator
}

func NewService(
	storage storage,
	tokenGenerator *token.Generator,
) *Service {
	return &Service{
		storage:   storage,
		generator: tokenGenerator,
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

	EncryptedPassword, err := secure.EncryptPassword(password)
	if err != nil {
		return FailedToEncryptPassword
	}

	user := user.User{
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

	if secure.ComparePasswords(user.Password, password) {
		return "", InvalidCredentials
	}

	return s.generator.Generate(login)
}

func (s *Service) ValidateToken(tokenString string) (string, error) {
	return s.generator.Validate(tokenString)
}
