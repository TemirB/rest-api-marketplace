package post

import (
	"github.com/TemirB/rest-api-marketplace/internal/pkg/models/post"
	postStorage "github.com/TemirB/rest-api-marketplace/internal/storage/post"
)

type storage interface {
	Create(post *post.Post) error
	GetByID(id uint) (*post.Post, error)
	GetAll(postStorage.SortParams, postStorage.FilterParams) ([]*post.Post, error)
}

type Service struct {
	repository storage
}

func NewService(repository storage) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) CreatePost(post *post.Post) (*post.Post, error) {
	err := s.repository.Create(post)
	if err != nil {
		return nil, err
	}

	return post, nil
}

func (s *Service) GetPosts(sort postStorage.SortParams, filter postStorage.FilterParams) ([]*post.Post, error) {
	posts, err := s.repository.GetAll(sort, filter)
	if err != nil {
		return nil, err
	}

	return posts, nil
}
