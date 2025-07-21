package post

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"
)

var (
	ErrMissingFields = fmt.Errorf("title and description are required")
	ErrTooLong       = fmt.Errorf("title must not exceed 100 characters, description must not exceed 2000 characters")
	ErrNegativePrice = fmt.Errorf("price must be a positive number")
	ErrRequiredURL   = fmt.Errorf("image URL is required")
)

var imageExts = map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true}

func validatePost(post *Post) error {
	if post.Title == "" || post.Description == "" {
		return ErrMissingFields
	}
	if len(post.Title) > 100 || len(post.Description) > 2000 {
		return ErrTooLong
	}
	if post.Price < 0 {
		return ErrNegativePrice
	}

	u, err := url.ParseRequestURI(post.ImageURL)
	if err != nil {
		return errors.New("invalid image URL")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("invalid URL scheme")
	}
	ext := strings.ToLower(path.Ext(u.Path))
	if !imageExts[ext] {
		return errors.New("unsupported image extension")
	}
	return nil
}
