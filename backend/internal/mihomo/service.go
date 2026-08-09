package mihomo

import (
	"fmt"
	"time"

	"github.com/dzx941/3m-ui/backend/internal/config"
)

type Service struct {
	pm *ProcessManager
	cm *ConfigManager
}

var GlobalService *Service

func InitService(cfg *config.Config) {
	pm := GetProcessManager(cfg.Mihomo.Binary, cfg.Mihomo.Config)
	cm := NewConfigManager(cfg.Mihomo.Config)
	GlobalService = &Service{
		pm: pm,
		cm: cm,
	}
}

func (s *Service) StartMihomo() error {
	if s == nil || s.pm == nil {
		return fmt.Errorf("mihomo service not initialized")
	}
	return s.pm.Start()
}

func (s *Service) StopMihomo() error {
	if s == nil || s.pm == nil {
		return fmt.Errorf("mihomo service not initialized")
	}
	return s.pm.Stop()
}

func (s *Service) RestartMihomo() error {
	if s == nil || s.pm == nil {
		return fmt.Errorf("mihomo service not initialized")
	}
	return s.pm.Restart()
}

func (s *Service) GetStatus() (*StatusResponse, error) {
	if s == nil || s.pm == nil {
		return nil, fmt.Errorf("mihomo service not initialized")
	}
	return s.pm.Status()
}

func (s *Service) GetLogs() ([]LogResponse, error) {
	// Mock implementation for Phase 2 placeholder logs API
	return []LogResponse{
		{
			Timestamp: time.Now().Add(-5 * time.Second),
			Level:     "info",
			Payload:   "Mihomo Core initialized successfully.",
		},
		{
			Timestamp: time.Now().Add(-2 * time.Second),
			Level:     "info",
			Payload:   "Start listening to inbound ports configured.",
		},
	}, nil
}
