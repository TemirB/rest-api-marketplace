package post

type Service struct {
	repository *storage
}

func NewService(repository *storage) *Service {
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

func (s *Service) GetPosts(sort *SortParams, filter *FilterParams) ([]*Post, error) {
	return s.repository.GetAll(sort, filter)
}
