package service

type Service struct {
}

func New() (s *Service) {
	s = &Service{}
	if err := s.init(); err != nil {
		panic(err)
	}
	return s
}

func (s *Service) init() error {

	return nil
}
