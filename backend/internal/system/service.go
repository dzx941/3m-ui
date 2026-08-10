package system

type Service struct{}

var GlobalService *Service

func InitService() {
	GlobalService = &Service{}
}

func (s *Service) GetStatus() *SystemStats {
	return GetSystemStats()
}
