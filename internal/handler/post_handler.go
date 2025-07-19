package handler

import (
	"github.com/TemirB/rest-api-marketplace/internal/post"
)

type PostHandler interface {
	CreatePost(post post.Post) (string, error)
}
