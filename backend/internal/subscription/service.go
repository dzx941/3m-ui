package subscription

import (
	"crypto/rand"
	"encoding/hex"
	"errors"

	"github.com/dzx941/3m-ui/backend/internal/database/models"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func GenerateToken() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (s *Service) Create(userID uint, format string) (*models.Subscription, error) {
	if userID == 0 {
		return nil, errors.New("invalid user id")
	}

	token, err := GenerateToken()
	if err != nil {
		return nil, err
	}

	sub := &models.Subscription{
		UserID: userID,
		Token: token,
		Format: format,
	}

	if err := s.db.Create(sub).Error; err != nil {
		return nil, err
	}

	return sub, nil
}
