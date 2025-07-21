package post

import (
	"go.uber.org/zap"
)

// mockgen  -source=service.go -destination=service_mock_test.go -package=post

type storage interface {
	Create(post *Post) error
	GetAll(sort *SortParams, filter *FilterParams) ([]*Post, error)
	Update(post *Post) error
	Delete(id uint64) error
	GetByID(id uint) (*Post, error)
}

type Service struct {
	repository storage
	logger     *zap.Logger
}

func NewService(repository storage, logger *zap.Logger) *Service {
	return &Service{
		repository: repository,
		logger:     logger,
	}
}

func (s *Service) CreatePost(post *Post) (*Post, error) {
	err := validatePost(post)
	if err != nil {
		s.logger.Error(
			"Validation error",
			zap.Error(err),
		)
		return nil, err
	}
	err = s.repository.Create(post)
	if err != nil {
		return nil, err
	}

	return post, nil
}

func (s *Service) GetPosts(sort *SortParams, filter *FilterParams) ([]*Post, error) {
	return s.repository.GetAll(sort, filter)
}

func (s *Service) DeletePost(id uint64) error {
	return s.repository.Delete(id)
}

func (s *Service) UpdatePost(post *Post) error {
	err := validatePost(post)
	if err != nil {
		s.logger.Error(
			"Validation error",
			zap.Error(err),
		)
		return err
	}

	return s.repository.Update(post)
}

func (s *Service) GetPostByID(id uint) (*Post, error) {
	return s.repository.GetByID(id)
}
