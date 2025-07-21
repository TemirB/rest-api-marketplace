package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUser(t *testing.T) {
	testCases := []struct {
		name        string
		description string

		login    string
		password string
	}{
		{
			name:        "1.Valid_User",
			description: "Create a new user with valid login and password",

			login:    "testuser",
			password: "securepassword",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			user := NewUser(tc.login, tc.password)
			assert.Equal(t, tc.login, user.Login, "User login should match")
			assert.Equal(t, tc.password, user.Password, "User password should match")
		})
	}
}
