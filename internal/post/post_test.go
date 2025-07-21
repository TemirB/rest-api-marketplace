package post

import (
	"testing"

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
		name        string
		description string

		postOpts opts
	}{
		{
			name:        "1.Valid_Post",
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
		t.Run(tc.name, func(t *testing.T) {
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
