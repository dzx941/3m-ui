package traffic

import (
	"time"

	"github.com/dzx941/3m-ui/backend/internal/database/models"
	"gorm.io/gorm"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

func (s *UserService) AddSample(userID uint, up, down int64, online bool) error {
	record := &models.TrafficRecord{
		ProxyUserID: userID,
		UploadBytes: up,
		DownloadBytes: down,
		Online: online,
	}
	if err := s.db.Create(record).Error; err != nil {
		return err
	}

	return s.db.Model(&models.ProxyUser{}).
		Where("id = ?", userID).
		Updates(map[string]any{
			"traffic_used": gorm.Expr("traffic_used + ?", up+down),
		}).Error
}

func IsExpired(u models.ProxyUser) bool {
	return !u.ExpireTime.IsZero() && u.ExpireTime.Before(time.Now())
}
