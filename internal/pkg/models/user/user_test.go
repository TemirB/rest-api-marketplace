package models

import (
	"testing"

	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/runner"
	"github.com/stretchr/testify/assert"
)

func TestNewUser(t *testing.T) {
	testCases := []struct {
		id          string
		description string

		login    string
		password string
	}{
		{
			id:          "1.Valid_User",
			description: "Create a new user with valid login and password",

			login:    "testuser",
			password: "securepassword",
		},
	}

	for _, tc := range testCases {
		runner.Run(t, tc.id, func(tp provider.T) {
			tp.Story("NewUser")

			tp.ID(tc.id)
			tp.Description(tc.description)

			user := NewUser(tc.login, tc.password)
			assert.Equal(t, tc.login, user.Login, "User login should match")
			assert.Equal(t, tc.password, user.Password, "User password should match")
		})
	}
}
