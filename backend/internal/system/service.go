package system

type Service struct{}

// NewService constructs the system service explicitly.
func NewService() *Service { return &Service{} }

func (s *Service) GetStatus() *SystemStats {
	return GetSystemStats()
}
