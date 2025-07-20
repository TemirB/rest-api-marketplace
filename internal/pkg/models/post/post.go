package post

import "time"

type Post struct {
	ID          uint    `gorm:"primaryKey"`
	Title       string  `gorm:"size:200"`
	Description string  `gorm:"type:text"`
	Price       float64 `gorm:"type:numeric(10,2)"`
	ImageURL    string  `gorm:"size:500"`

	CreatedAt time.Time `gorm:"index"`
	Owner     string    // References User.Login
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
