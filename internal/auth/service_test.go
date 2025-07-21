package auth

import (
	"database/sql"
	"testing"

	"github.com/TemirB/rest-api-marketplace/pkg/hash"
	gomock "github.com/golang/mock/gomock"
	"go.uber.org/zap"
)

func TestService_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testCases := []struct {
		testID string
		name   string

		login      string
		password   string
		setupMocks func(ctrl *gomock.Controller) *Service

		expectedError error
	}{
		{
			name: "Valid_Registration",

			login:    "testuser",
			password: "securepassword",

			setupMocks: func(ctrl *gomock.Controller) *Service {
				storage := NewMockstorage(ctrl)
				manager := NewMockmanager(ctrl)

				service := NewService(storage, manager, zap.NewNop())

				storage.EXPECT().Exists("testuser").Return(false, nil)
				storage.EXPECT().Create(gomock.Any()).Return(nil)

				return service
			},

			expectedError: nil,
		},
		{
			name: "Registration_invalid_login",

			login: "",

			setupMocks: func(ctrl *gomock.Controller) *Service {
				service := NewService(nil, nil, zap.NewNop())

				return service
			},

			expectedError: ErrInvalidLogin,
		},
		{
			name: "Registration_invalid_password",

			login:    "testuser",
			password: "weak",

			setupMocks: func(ctrl *gomock.Controller) *Service {
				service := NewService(nil, nil, zap.NewNop())

				return service
			},

			expectedError: ErrInvalidPassword,
		},
		{
			name: "Registration_user_already_exists",

			login:    "testuser",
			password: "securepassword",

			setupMocks: func(ctrl *gomock.Controller) *Service {
				storage := NewMockstorage(ctrl)
				manager := NewMockmanager(ctrl)

				service := NewService(storage, manager, zap.NewNop())

				storage.EXPECT().Exists("testuser").Return(true, sql.ErrConnDone)

				return service
			},

			expectedError: sql.ErrConnDone,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := tc.setupMocks(ctrl)
			err := service.Register(tc.login, tc.password)

			if err != nil && tc.expectedError == nil {
				t.Fatalf("Expected no error, but got: %v", err)
			}
			if err == nil && tc.expectedError != nil {
				t.Fatalf("Expected error: %v, but got: nil", tc.expectedError)
			}
		})
	}
}

func TestService_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	securePassword, _ := hash.EncryptPassword("securepassword")

	testCases := []struct {
		name string

		login      string
		password   string
		setupMocks func(ctrl *gomock.Controller) *Service

		expectedError error
	}{
		{
			name: "Valid_Login",

			login:    "testuser",
			password: "securepassword",

			setupMocks: func(ctrl *gomock.Controller) *Service {
				storage := NewMockstorage(ctrl)
				manager := NewMockmanager(ctrl)

				service := NewService(storage, manager, zap.NewNop())

				storage.EXPECT().GetByLogin("testuser").Return(&User{Login: "testuser", Password: securePassword}, nil)
				manager.EXPECT().GenerateToken("testuser").Return("testToken", nil)

				return service
			},

			expectedError: nil,
		},
		{
			name: "Login_invalid_login",

			setupMocks: func(ctrl *gomock.Controller) *Service {
				service := NewService(nil, nil, zap.NewNop())

				return service
			},

			expectedError: ErrInvalidLogin,
		},
		{
			name: "Login_failed_to_find_user",

			login:    "testuser",
			password: "securepassword",

			setupMocks: func(ctrl *gomock.Controller) *Service {
				storage := NewMockstorage(ctrl)
				storage.EXPECT().GetByLogin("testuser").Return(nil, sql.ErrNoRows)
				service := NewService(storage, nil, zap.NewNop())

				return service
			},

			expectedError: ErrInvalidCredentials,
		},
		{
			name: "Login_failed_to_compare_passwords",

			login:    "testuser",
			password: "wrongpassword",

			setupMocks: func(ctrl *gomock.Controller) *Service {
				storage := NewMockstorage(ctrl)
				storage.EXPECT().GetByLogin("testuser").Return(&User{Login: "testuser", Password: securePassword}, nil)
				service := NewService(storage, nil, zap.NewNop())

				return service
			},

			expectedError: ErrInvalidCredentials,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := tc.setupMocks(ctrl)
			token, err := service.Login(tc.login, tc.password)

			if err != nil && tc.expectedError == nil {
				t.Fatalf("Expected no error, but got: %v", err)
			}
			if err == nil && tc.expectedError != nil {
				t.Fatalf("Expected error: %v, but got: nil", tc.expectedError)
			}

			if err == nil && token != "testToken" {
				t.Fatalf("Expected token 'testToken', but got: %s", token)
			}
		})
	}
}
