package auth

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestStorage_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	query := `INSERT INTO users (login, password) VALUES ($1, $2)`

	testCases := []struct {
		name string

		user       *User
		setupMocks func(ctrl *gomock.Controller) *Storage

		errorExpected bool
	}{
		{
			name: "Valid_User",

			user: &User{
				Login:    "testuser",
				Password: "securepassword",
			},

			setupMocks: func(ctrl *gomock.Controller) *Storage {
				MockRepository := NewMockRepository(ctrl)

				MockRepository.EXPECT().Exec(query, "testuser", "securepassword").Return(nil, nil)
				return NewStorage(MockRepository, zap.NewNop())
			},

			errorExpected: false,
		},
		{
			name: "Failed_Create",

			user: &User{
				Login:    "testuser",
				Password: "securepassword",
			},

			setupMocks: func(ctrl *gomock.Controller) *Storage {
				MockRepository := NewMockRepository(ctrl)

				MockRepository.EXPECT().Exec(query, "testuser", "securepassword").Return(nil, fmt.Errorf("failed to create user"))
				return NewStorage(MockRepository, zap.NewNop())
			},

			errorExpected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			storage := tc.setupMocks(ctrl)

			err := storage.Create(tc.user)
			if tc.errorExpected {
				assert.Error(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

// Подобное тестирование для Exists, GetByLogin, DeleteUser
