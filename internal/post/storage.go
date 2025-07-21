package post

// mockgen  -source=storage.go -destination=storage_mock_test.go -package=post

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var ErrPostNotFound = errors.New("post not found")

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
	const query = `
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
		&post.Owner,
		&post.CreatedAt,
	)
	if err != nil {
		r.logger.Error(
			"Failed to get post by ID",
			zap.Uint("id", id),
			zap.Error(err),
		)
		if err == sql.ErrNoRows {
			return nil, ErrPostNotFound
		}
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
	if filter == nil {
		filter = &FilterParams{MinPrice: 0, MaxPrice: -1}
	}
	if sort == nil {
		sort = &SortParams{Field: "created_at", Direction: "DESC"}
	}

	var (
		sb   strings.Builder
		args []interface{}
		idx  = 1
	)
	sb.WriteString("SELECT id, title, description, price, image_url, owner, created_at FROM posts WHERE 1=1")

	if filter.MaxPrice >= 0 {
		sb.WriteString(fmt.Sprintf(" AND price BETWEEN $%d AND $%d", idx, idx+1))
		args = append(args, filter.MinPrice, filter.MaxPrice)
		idx += 2
	} else {
		sb.WriteString(fmt.Sprintf(" AND price >= $%d", idx))
		args = append(args, filter.MinPrice)
		idx++
	}

	/*
		// По условию нужно помечать посты, которые принадлежат определенному владельцу, но можно и возвращать только свои посты
		   if filter.Owner != "" {
		       sb.WriteString(fmt.Sprintf(" AND owner = $%d", idx))
		       args = append(args, filter.Owner)
		       idx++
		   }
	*/

	sb.WriteString(fmt.Sprintf(" ORDER BY %s %s", sort.Field, sort.Direction))

	rows, err := r.repository.Query(sb.String(), args...)
	if err != nil {
		r.logger.Error("Failed to get posts",
			zap.Error(err),
			zap.Any("sort", sort),
			zap.Any("filter", filter))
		return nil, errors.Wrap(err, "failed to get posts")
	}
	defer rows.Close()

	var posts []*Post
	for rows.Next() {
		var p Post
		if err := rows.Scan(
			&p.ID,
			&p.Title,
			&p.Description,
			&p.Price,
			&p.ImageURL,
			&p.Owner,
			&p.CreatedAt,
		); err != nil {
			r.logger.Error("Failed to scan post row", zap.Error(err))
			continue
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

func (r *Storage) Update(post *Post) error {
	query := `
        UPDATE posts
        SET title=$1, description=$2, price=$3, image_url=$4
        WHERE id=$5
    `
	_, err := r.repository.Exec(query, post.Title, post.Description, post.Price, post.ImageURL, post.ID)
	if err != nil {
		r.logger.Error(
			"Failed to update post",
			zap.Uint("id", post.ID),
			zap.Error(err),
		)
		return errors.Errorf("failed to update post: %d", err)
	}
	return nil
}

func (r *Storage) Delete(id uint64) error {
	query := `DELETE FROM posts WHERE id = $1`
	_, err := r.repository.Exec(query, id)
	if err != nil {
		r.logger.Error(
			"Failed to delete post",
			zap.Uint64("id", id),
			zap.Error(err),
		)
		return errors.Errorf("failed to delete post: %d", err)
	}
	return nil
}
