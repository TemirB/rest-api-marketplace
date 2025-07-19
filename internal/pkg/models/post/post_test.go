package models

import (
	"testing"

	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/runner"
	"github.com/stretchr/testify/assert"
)

func TestNewPost(t *testing.T) {
	type opts struct {
		title       string
		description string
		price       float64
		imageURL    string
		ownerLogin  string
	}
	testCases := []struct {
		id          string
		description string

		postOpts opts
	}{
		{
			id:          "1.Valid_Post",
			description: "Create a new post with valid parameters",

			postOpts: opts{
				title:       "Test Post",
				description: "This is a test post description.",
				price:       99.99,
				imageURL:    "http://example.com/image.jpg",
				ownerLogin:  "testuser",
			},
		},
	}

	for _, tc := range testCases {
		runner.Run(t, tc.id, func(tp provider.T) {
			tp.Story("NewPost")

			tp.ID(tc.id)
			tp.Description(tc.description)

			post := NewPost(
				tc.postOpts.title,
				tc.postOpts.description,
				tc.postOpts.price,
				tc.postOpts.imageURL,
				tc.postOpts.ownerLogin,
			)

			assert.Equal(t, tc.postOpts.title, post.Title, "Post title should match")
			assert.Equal(t, tc.postOpts.description, post.Description, "Post description should match")
			assert.Equal(t, tc.postOpts.price, post.Price, "Post price should match")
			assert.Equal(t, tc.postOpts.imageURL, post.ImageURL, "Post image URL should match")
			assert.Equal(t, tc.postOpts.ownerLogin, post.Owner, "Post owner login should match")
		})
	}
}
