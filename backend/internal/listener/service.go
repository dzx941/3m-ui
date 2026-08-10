package listener

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dzx941/3m-ui/backend/internal/database/models"
	"gorm.io/gorm"
)

type Service struct {
	db         *gorm.DB
	configPath string
}

var GlobalService *Service

func InitService(db *gorm.DB, configPath string) {
	GlobalService = &Service{
		db:         db,
		configPath: configPath,
	}
}

func (s *Service) Create(l *models.Listener) error {
	if err := s.db.Create(l).Error; err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}
	return s.RegenerateConfig()
}

func (s *Service) GetAll() ([]models.Listener, error) {
	var list []models.Listener
	if err := s.db.Find(&list).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch listeners: %w", err)
	}
	return list, nil
}

func (s *Service) GetByID(id uint) (*models.Listener, error) {
	var l models.Listener
	if err := s.db.First(&l, id).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch listener by id %d: %w", id, err)
	}
	return &l, nil
}

func (s *Service) Update(l *models.Listener) error {
	if err := s.db.Save(l).Error; err != nil {
		return fmt.Errorf("failed to update listener: %w", err)
	}
	return s.RegenerateConfig()
}

func (s *Service) Delete(id uint) error {
	if err := s.db.Delete(&models.Listener{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete listener: %w", err)
	}
	return s.RegenerateConfig()
}

// RegenerateConfig fetches all listeners, runs generation, and saves to file
func (s *Service) RegenerateConfig() error {
	listeners, err := s.GetAll()
	if err != nil {
		return err
	}

	yamlContent, err := GenerateConfigYAML(listeners)
	if err != nil {
		return err
	}

	// Create directories if missing
	dir := filepath.Dir(s.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory %s: %w", dir, err)
	}

	// Save to file
	if err := os.WriteFile(s.configPath, []byte(yamlContent), 0644); err != nil {
		return fmt.Errorf("failed to write config.yaml: %w", err)
	}

	return nil
}

// TriggerReload forces manual regeneration and signals reloading logic
func (s *Service) TriggerReload(id uint) error {
	// First, regenerate the file
	if err := s.RegenerateConfig(); err != nil {
		return err
	}

	// Then, trigger reload in Mihomo Core
	// In the future, we would call the Mihomo REST API /configs?force=true
	return nil
}
