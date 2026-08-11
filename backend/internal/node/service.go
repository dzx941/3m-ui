package node

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/dzx941/3m-ui/backend/internal/database/models"
	"github.com/dzx941/3m-ui/backend/internal/mihomo"
	mihomoConfig "github.com/dzx941/3m-ui/backend/internal/mihomo/config"
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
	if err := ValidateNode(l); err != nil {
		return err
	}

	if err := s.db.Create(l).Error; err != nil {
		return fmt.Errorf("failed to create node: %w", err)
	}

	return s.RegenerateConfig()
}

func (s *Service) GetAll() ([]models.Listener, error) {
	var list []models.Listener
	if err := s.db.Find(&list).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch nodes: %w", err)
	}
	return list, nil
}

func (s *Service) GetByID(id uint) (*models.Listener, error) {
	var l models.Listener
	if err := s.db.First(&l, id).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch node by id %d: %w", id, err)
	}
	return &l, nil
}

func (s *Service) Update(l *models.Listener) error {
	if err := ValidateNode(l); err != nil {
		return err
	}

	if err := s.db.Save(l).Error; err != nil {
		return fmt.Errorf("failed to update node: %w", err)
	}

	return s.RegenerateConfig()
}

func (s *Service) Delete(id uint) error {
	if err := s.db.Delete(&models.Listener{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete node: %w", err)
	}

	return s.RegenerateConfig()
}

// RegenerateConfig regenerates the complete Mihomo configuration through the
// Config Engine. Node services never write a partial listeners-only YAML file.
func (s *Service) RegenerateConfig() error {
	if GlobalService == nil {
		return fmt.Errorf("node service is not initialized")
	}
	engine := mihomoConfig.NewConfigEngine(s.db)
	yamlContent, err := engine.GenerateFinalConfig()
	if err != nil {
		return err
	}
	dir := filepath.Dir(s.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}
	if err := os.WriteFile(s.configPath, []byte(yamlContent), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	// Best-effort hot reload: ask the running Mihomo Core to reload the
	// config we just wrote via its external controller API. If Mihomo isn't
	// running or the controller is unreachable, this is not a failure of
	// config regeneration itself (the file on disk is still correct and
	// will be picked up on the next manual/automatic start), so we only log.
	tmpl := mihomoConfig.GetDefaultTemplate()
	controllerURL := "http://" + tmpl.ExternalController
	if err := mihomo.NewExternalControllerAPI(controllerURL, tmpl.Secret).ReloadConfig(map[string]interface{}{
		"path": s.configPath,
	}); err != nil {
		log.Printf("node: mihomo hot reload skipped (core unreachable): %v", err)
	}

	return nil
}

// TriggerReload forces manual config regeneration and triggers hot reloading
func (s *Service) TriggerReload(id uint) error {
	return s.RegenerateConfig()
}
