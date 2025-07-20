package post

import "time"

type Post struct {
	ID          uint
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Price       float64 `json:"price"` // float можно заменить на decimal для большей точности, но для простоты оставим float
	ImageURL    string  `json:"image_url"`

	CreatedAt time.Time `json:"created_at"`
	Owner     string    `json:"owner"`
	IsOwner   bool      `json:"is_owner"`
}

func NewPost(
	title,
	description string,
	price float64,
	imageURL, ownerLogin string,
) *Post {
	return &Post{
		Title:       title,
		Description: description,
		Price:       price,
		ImageURL:    imageURL,
		Owner:       ownerLogin,
	}
}

type SortParams struct {
	Field     string // "price" | "created_at"
	Direction string // "asc" | "desc"
}

type FilterParams struct {
	MinPrice float64
	MaxPrice float64
	Owner    string
}
