package post

import "fmt"

var (
	ErrMissingFields = fmt.Errorf("title and description are required")
	ErrTooLong       = fmt.Errorf("title must not exceed 100 characters, description must not exceed 2000 characters")
	ErrNegativePrice = fmt.Errorf("price must be a positive number")
	ErrRequiredURL   = fmt.Errorf("image URL is required")
)

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
	if post.ImageURL == "" {
		// Вообще тут стоит добавить валидацию URL, но пока просто проверяем, что он не пустой
		return ErrRequiredURL
	}

	return nil
}
