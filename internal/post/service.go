package post

import "go.uber.org/zap"

// mockgen  -source=service.go -destination=service_mock_test.go -package=post

type storage interface {
	Create(post *Post) error
	GetAll(sort *SortParams, filter *FilterParams) ([]*Post, error)
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
