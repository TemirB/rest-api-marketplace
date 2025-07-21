package post

// mockgen  -source=storage.go -destination=storage_mock_test.go -package=post

import (
	"database/sql"
	"fmt"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Repository interface {
	Exec(query string, args ...any) (sql.Result, error)
	QueryRow(query string, args ...any) *sql.Row
	Query(query string, args ...any) (*sql.Rows, error)
}

type Storage struct {
	repository Repository
	logger     *zap.Logger
}

func NewStorage(repository Repository, logger *zap.Logger) *Storage {
	return &Storage{
		repository: repository,
		logger:     logger,
	}
}

func (r *Storage) Create(post *Post) error {
	query := `
		INSERT INTO posts (title, description, price, image_url, owner)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`
	err := r.repository.QueryRow(
		query,
		post.Title,
		post.Description,
		post.Price,
		post.ImageURL,
		post.Owner,
	).Scan(&post.ID, &post.CreatedAt)

	if err != nil {
		r.logger.Error(
			"Failed to create post",
			zap.Error(err),
		)
		return errors.Errorf("failed to create post: %d", err)
	}
	return nil
}

func (r *Storage) GetByID(id uint) (*Post, error) {
	query := `
		SELECT id, title, description, price, image_url, owner, created_at
		FROM posts WHERE id = $1
	`
	row := r.repository.QueryRow(query, id)

	var post Post
	err := row.Scan(
		&post.ID,
		&post.Title,
		&post.Description,
		&post.Price,
		&post.ImageURL,
		&post.CreatedAt,
		&post.Owner,
	)
	if err != nil {
		r.logger.Error(
			"Failed to get post by ID",
			zap.Uint("id", id),
			zap.Error(err),
		)
		return nil, errors.Errorf("failed to get post: %d", err)
	}

	return &post, nil
}

func setField(query string, sort *SortParams) string {
	switch sort.Field {
	case "price":
		query += fmt.Sprintf(" ORDER BY price %s", sort.Direction)
	case "created_at":
		query += fmt.Sprintf(" ORDER BY created_at %s", sort.Direction)
	default:
		query += " ORDER BY created_at DESC"
	}
	return query
}

func (r *Storage) GetAll(sort *SortParams, filter *FilterParams) ([]*Post, error) {
	var query string
	var rows *sql.Rows
	var err error
	if filter.MaxPrice < 0 {
		query = fmt.Sprintf("SELECT ... FROM posts WHERE price >= $1 ORDER BY %s %s", sort.Field, sort.Direction)
		query += setField(query, sort)
		rows, err = r.repository.Query(query, filter.MinPrice)
	} else {
		query = fmt.Sprintf("SELECT ... FROM posts WHERE price BETWEEN $1 AND $2 ORDER BY %s %s", sort.Field, sort.Direction)
		query += setField(query, sort)
		rows, err = r.repository.Query(query, filter.MinPrice, filter.MaxPrice)
	}
	if err != nil {
		r.logger.Error(
			"Failed to get posts",
			zap.Error(err),
			zap.Any("sort", sort),
			zap.Any("filter", filter))
		return nil, errors.Wrap(err, "failed to get posts")
	}
	defer rows.Close()

	var posts []*Post
	for rows.Next() {
		var p Post
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
			r.logger.Error("Failed to scan post row", zap.Error(err))
			continue // Пропускаем проблемные строки или return nil, errors.Wrap(...)
		}
		if filter.Owner != "" && p.Owner == filter.Owner {
			p.IsOwner = true
		}
		posts = append(posts, &p)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Error iterating over posts", zap.Error(err))
		return nil, errors.Wrap(err, "error iterating over posts")
	}

	return posts, nil
}

func (r *Storage) Delete(id uint) error {
	query := `DELETE FROM posts WHERE id = $1`
	_, err := r.repository.Exec(query, id)
	if err != nil {
		r.logger.Error("Failed to delete post", zap.Uint("id", id), zap.Error(err))
		return errors.Errorf("failed to delete post: %d", err)
	}
	return nil
}
