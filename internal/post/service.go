package post

import (
	postStorage "github.com/TemirB/rest-api-marketplace/internal/database/post"
)

type storage interface {
	Create(post *Post) error
	GetByID(id uint) (*Post, error)
	GetAll(postStorage.SortParams, postStorage.FilterParams) ([]*Post, error)
}

type Service struct {
	repository storage
}

func NewService(repository storage) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) CreatePost(post *Post) (*Post, error) {
	err := s.repository.Create(post)
	if err != nil {
		return nil, err
	}

	return post, nil
}

func (s *Service) GetPosts(sort postStorage.SortParams, filter postStorage.FilterParams) ([]*Post, error) {
	posts, err := s.repository.GetAll(sort, filter)
	if err != nil {
		return nil, err
	}

	return posts, nil
}
