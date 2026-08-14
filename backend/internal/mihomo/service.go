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

// NewService constructs a Mihomo service with explicit dependencies.
func NewService(cfg *config.Config) *Service {
	return &Service{
		pm: GetProcessManager(cfg.Mihomo.Binary, cfg.Mihomo.Config),
		cm: NewConfigManager(cfg.Mihomo.Config),
	}
}

func (s *Service) StartMihomo() error {
	if s == nil || s.pm == nil {
		return fmt.Errorf("mihomo service not initialized")
	}
	if err := s.pm.ValidateConfig(); err != nil {
		return err
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

func (s *Service) SaveConfig(content string) error {
	if s == nil || s.cm == nil || s.pm == nil {
		return fmt.Errorf("mihomo service not initialized")
	}
	old, readErr := s.cm.ReadConfig()
	if err := s.cm.SaveConfig(content); err != nil {
		return err
	}
	if err := s.pm.ValidateConfig(); err != nil {
		if readErr == nil {
			_ = s.cm.SaveConfig(old)
		}
		return err
	}
	return nil
}

func (s *Service) ApplyConfig(content string) error {
	if s == nil || s.pm == nil || s.cm == nil {
		return fmt.Errorf("mihomo service not initialized")
	}
	old, readErr := s.cm.ReadConfig()
	wasRunning := s.pm.IsRunning()
	if err := s.cm.SaveConfig(content); err != nil {
		return err
	}
	if err := s.pm.ValidateConfig(); err != nil {
		if readErr == nil {
			_ = s.cm.SaveConfig(old)
		}
		return err
	}
	if !wasRunning {
		return s.pm.Start()
	}
	if err := s.pm.Restart(); err != nil {
		if readErr == nil {
			if restoreErr := s.cm.SaveConfig(old); restoreErr == nil {
				_ = s.pm.Restart()
			}
		}
		return err
	}
	return nil
}

func (s *Service) GetLogs() ([]LogResponse, error) {
	if s == nil || s.pm == nil {
		return nil, fmt.Errorf("mihomo service not initialized")
	}
	lines := s.pm.Logs()
	result := make([]LogResponse, 0, len(lines))
	for _, line := range lines {
		result = append(result, LogResponse{Timestamp: time.Now(), Level: "info", Payload: line})
	}
	return result, nil
}
