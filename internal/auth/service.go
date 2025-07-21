package auth

// mockgen  -source=service.go -destination=service_mock_test.go -package=auth

import (
	"errors"
	"fmt"

	"github.com/TemirB/rest-api-marketplace/pkg/hash"
	"go.uber.org/zap"
)

var (
	ErrInvalidLogin            = errors.New("login must be 3-50 characters")
	ErrInvalidPassword         = errors.New("password must be at least 8 characters")
	ErrFailedToEncryptPassword = errors.New("failed to encrypt password")

	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrWrongPassword      = errors.New("wrong password")

	ErrUserExists = errors.New("user already exists")
)

type storage interface {
	Create(user *User) error
	GetByLogin(login string) (*User, error)
	Exists(login string) (bool, error)
}
type manager interface {
	GenerateToken(login string) (string, error)
	ValidateToken(tokenStr string) (string, error)
}

type Service struct {
	storage storage
	manager manager
	logger  *zap.Logger
}

func NewService(
	storage storage,
	manager manager,
	logger *zap.Logger,
) *Service {
	return &Service{
		storage: storage,
		manager: manager,
		logger:  logger,
	}
}

func (s *Service) Register(login, password string) error {
	if !validLogin(login) {
		s.logger.Info(
			"invalid login",
			zap.String("login", login),
		)
		return ErrInvalidLogin
	}

	if !validPassword(password) {
		s.logger.Info(
			"invalid password",
			// zap.String("password", password), is it legit???
		)
		return ErrInvalidPassword
	}

	exists, err := s.userExists(login)
	if err != nil {
		return fmt.Errorf("unable to check user existence: %w", err)
	}
	if exists {
		return ErrUserExists
	}

	EncryptedPassword, err := hash.EncryptPassword(password)
	if err != nil {
		s.logger.Error(
			"failed to encrypt password",
			zap.String("login", login),
			zap.Error(err),
		)
		return ErrFailedToEncryptPassword
	}

	user := User{
		Login:    login,
		Password: string(EncryptedPassword),
	}

	return s.storage.Create(&user)
}

func (s *Service) Login(login, password string) (string, error) {
	if !validLogin(login) {
		s.logger.Info(
			"invalid login",
			zap.String("login", login),
		)
		return "", ErrInvalidLogin
	}
	user, err := s.storage.GetByLogin(login)
	if err != nil {
		s.logger.Error(
			"failed to get user by login",
			zap.String("login", login),
			zap.Error(err),
		)
		return "", ErrInvalidCredentials
	}

	if !hash.ComparePasswords(user.Password, password) {
		s.logger.Info(
			"password is incorrect",
			zap.String("login", login),
			// zap.String("password", password), is it legit???
		)
		return "", ErrWrongPassword
	}

	return s.manager.GenerateToken(login)
}

func (s *Service) ValidateToken(tokenString string) (string, error) {
	return s.manager.ValidateToken(tokenString)
}
