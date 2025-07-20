package storage

import (
	"fmt"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/TemirB/rest-api-marketplace/internal/pkg/db"
	"github.com/TemirB/rest-api-marketplace/internal/pkg/models/post"
)

type storage struct {
	repository *db.Repository
}

func NewStorage(repository *db.Repository) *storage {
	return &storage{
		repository: repository,
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

func (r *storage) Create(post *post.Post) error {
	query := `
		INSERT INTO posts (title, description, price, image_url, owner)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`
	err := r.repository.DB.QueryRow(
		query,
		post.Title,
		post.Description,
		post.Price,
		post.ImageURL,
		post.Owner,
	).Scan(&post.ID, &post.CreatedAt)

	if err != nil {
		r.repository.Logger.Error("Failed to create post", zap.Error(err))
		return errors.Errorf("failed to create post: %d", err)
	}
	return nil
}

func (r *storage) GetByID(id uint) (*post.Post, error) {
	query := `
		SELECT id, title, description, price, image_url, owner, created_at
		FROM posts WHERE id = $1
	`
	row := r.repository.DB.QueryRow(query, id)

	var post post.Post
	err := row.Scan(
		&post.ID,
		&post.Title,
		&post.Description,
		&post.Price,
		&post.ImageURL,
		&post.Owner,
		&post.CreatedAt,
	)
	if err != nil {
		r.repository.Logger.Error(
			"Failed to get post by ID",
			zap.Uint("id", id),
			zap.Error(err),
		)
		return nil, errors.Errorf("failed to get post: %d", err)
	}

	return &post, nil
}

func (r *storage) GetAll(sort SortParams, filter FilterParams) ([]*post.Post, error) {
	query := `
		SELECT id, title, description, price, image_url, owner, created_at
		FROM posts
		WHERE ($1 = 0 OR price >= $1)
		AND ($2 = 0 OR price <= $2)
	`

	switch sort.Field {
	case "price":
		query += fmt.Sprintf(" ORDER BY price %s", sort.Direction)
	case "created_at":
		query += fmt.Sprintf(" ORDER BY created_at %s", sort.Direction)
	default:
		query += " ORDER BY created_at DESC"
	}

	rows, err := r.repository.DB.Query(query, filter.MinPrice, filter.MaxPrice)
	if err != nil {
		r.repository.Logger.Error("Failed to get posts",
			zap.Error(err),
			zap.Any("sort", sort),
			zap.Any("filter", filter))
		return nil, errors.Wrap(err, "failed to get posts")
	}
	defer rows.Close()

	var posts []*post.Post
	for rows.Next() {
		var p post.Post
		err := rows.Scan(
			&p.ID,
			&p.Title,
			&p.Description,
			&p.Price,
			&p.ImageURL,
			&p.Owner,
			&p.CreatedAt,
		)
		if err != nil {
			r.repository.Logger.Error("Failed to scan post row", zap.Error(err))
			continue // Пропускаем проблемные строки или return nil, errors.Wrap(...)
		}
		posts = append(posts, &p)
	}

	if err := rows.Err(); err != nil {
		r.repository.Logger.Error("Error iterating over posts", zap.Error(err))
		return nil, errors.Wrap(err, "error iterating over posts")
	}

	return posts, nil
}

func (r *storage) Delete(id uint) error {
	query := `DELETE FROM posts WHERE id = $1`
	_, err := r.repository.DB.Exec(query, id)
	if err != nil {
		r.repository.Logger.Error("Failed to delete post", zap.Uint("id", id), zap.Error(err))
		return errors.Errorf("failed to delete post: %d", err)
	}
	return nil
}
