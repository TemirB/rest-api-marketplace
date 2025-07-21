package auth

// mockgen  -source=service.go -destination=service_mock_test.go -package=auth

import (
	"errors"

	"github.com/TemirB/rest-api-marketplace/pkg/hash"
	"go.uber.org/zap"
)

var (
	InvalidLogin            = errors.New("login must be 3-50 characters")
	InvalidPassword         = errors.New("password must be at least 8 characters")
	UserAlreadyExists       = errors.New("user already exists")
	FailedToEncryptPassword = errors.New("failed to encrypt password")

	InvalidCredentials = errors.New("invalid credentials")
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
		return InvalidLogin
	}

	if !validPassword(password) {
		s.logger.Info(
			"invalid password",
			// zap.String("password", password), is it legit???
		)
		return InvalidPassword
	}

	if no, err := s.userExist(login); no || err != nil {
		s.logger.Info(
			"user already exists or db error",
			zap.String("login", login),
			zap.Error(err),
		)
		return err
	}

	EncryptedPassword, err := hash.EncryptPassword(password)
	if err != nil {
		s.logger.Error(
			"failed to encrypt password",
			zap.String("login", login),
			zap.Error(err),
		)
		return FailedToEncryptPassword
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
		return "", InvalidLogin
	}
	user, err := s.storage.GetByLogin(login)
	if err != nil {
		s.logger.Error(
			"failed to get user by login",
			zap.String("login", login),
			zap.Error(err),
		)
		return "", InvalidCredentials
	}

	if !hash.ComparePasswords(user.Password, password) {
		s.logger.Info(
			"password is incorrect",
			zap.String("login", login),
			// zap.String("password", password), is it legit???
		)
		return "", InvalidCredentials
	}

	return s.manager.GenerateToken(login)
}

func (s *Service) ValidateToken(tokenString string) (string, error) {
	return s.manager.ValidateToken(tokenString)
}
