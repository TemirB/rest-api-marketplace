package models_test

import (
	"testing"

	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/runner"
	"github.com/stretchr/testify/assert"

	models "user"
)

func TestNewUser(t *testing.T) {
	testCases := []struct {
		name        string
		description string

		login    string
		password string
	}{
		{
			name:        "Valid User",
			description: "Create a new user with valid login and password",

			login:    "testuser",
			password: "securepassword",
		},
	}

	for _, tc := range testCases {
		runner.Run(tc.name, func(tp *provider.T) {
			tp.Name(tc.name)
			tp.Description(tc.description)

			user := models.NewUser(tc.login, tc.password)
			assert.Equal(t, tc.login, user.Login, "User login should match")
			assert.Equal(t, tc.password, user.Password, "User password should match")
		})
	}
}
