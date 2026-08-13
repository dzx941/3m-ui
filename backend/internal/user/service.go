package user

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dzx941/3m-ui/backend/internal/database/models"
	"github.com/dzx941/3m-ui/backend/internal/node"
	"github.com/dzx941/3m-ui/backend/internal/security"
	"gorm.io/gorm"
)

type Service struct { db *gorm.DB }
var GlobalService *Service
func InitService(db *gorm.DB) { GlobalService = &Service{db: db} }

type CreateInput struct { Username string `json:"username" binding:"required"`; Password string `json:"password"`; UUID string `json:"uuid"`; TrafficLimit int64 `json:"traffic_limit"`; ExpireTime *time.Time `json:"expire_time"`; Enabled *bool `json:"enabled"` }
type UpdateInput struct { Username string `json:"username"`; Password string `json:"password"`; UUID string `json:"uuid"`; TrafficLimit *int64 `json:"traffic_limit"`; ExpireTime *time.Time `json:"expire_time"`; Enabled *bool `json:"enabled"` }
type Credential struct { Username, Password, UUID string }

func (s *Service) Create(in CreateInput) (*models.ProxyUser, error) {
	username := strings.TrimSpace(in.Username); if username == "" { return nil, errors.New("username is required") }
	password := in.Password; if password == "" { password, _ = randomToken(24) }
	uuid := in.UUID; if uuid == "" { var err error; uuid, err = newUUID(); if err != nil { return nil, err } }
	expire := time.Time{}; if in.ExpireTime != nil { expire = in.ExpireTime.UTC() }
	enabled := true; if in.Enabled != nil { enabled = *in.Enabled }
	encrypted, err := encryptPassword(password); if err != nil { return nil, err }
	u := &models.ProxyUser{Username: username, PasswordEncrypted: encrypted, UUID: uuid, TrafficLimit: in.TrafficLimit, ExpireTime: expire, Enabled: enabled}
	if err := s.db.Create(u).Error; err != nil { return nil, fmt.Errorf("create proxy user: %w", err) }; return u, nil
}
func (s *Service) GetAll() ([]models.ProxyUser, error) { var users []models.ProxyUser; if err := s.db.Order("id desc").Find(&users).Error; err != nil { return nil, err }; return users, nil }
func (s *Service) GetByID(id uint) (*models.ProxyUser, error) { var u models.ProxyUser; if err := s.db.First(&u, id).Error; err != nil { return nil, err }; return &u, nil }
func (s *Service) Update(id uint, in UpdateInput) (*models.ProxyUser, error) { u, err := s.GetByID(id); if err != nil { return nil, err }; if strings.TrimSpace(in.Username) != "" { u.Username = strings.TrimSpace(in.Username) }; if in.Password != "" { u.PasswordEncrypted, err = encryptPassword(in.Password); if err != nil { return nil, err } }; if in.UUID != "" { u.UUID = in.UUID }; if in.TrafficLimit != nil { u.TrafficLimit = *in.TrafficLimit }; if in.ExpireTime != nil { u.ExpireTime = in.ExpireTime.UTC() }; if in.Enabled != nil { u.Enabled = *in.Enabled }; if err := s.db.Save(u).Error; err != nil { return nil, fmt.Errorf("update proxy user: %w", err) }; return u, nil }
func (s *Service) Delete(id uint) error { return s.db.Transaction(func(tx *gorm.DB) error { if err := tx.Where("proxy_user_id = ?", id).Delete(&models.ListenerUser{}).Error; err != nil { return err }; return tx.Delete(&models.ProxyUser{}, id).Error }) }
func (s *Service) BindListeners(userID uint, listenerIDs []uint) error { return s.db.Transaction(func(tx *gorm.DB) error { var user models.ProxyUser; if err := tx.First(&user, userID).Error; err != nil { return err }; desired := make([]uint, 0, len(listenerIDs)); seen := map[uint]struct{}{}; for _, id := range listenerIDs { if _, ok := seen[id]; ok { continue }; seen[id] = struct{}{}; desired = append(desired, id) }; if len(desired) > 0 { var count int64; if err := tx.Model(&models.Listener{}).Where("id IN ?", desired).Count(&count).Error; err != nil { return err }; if count != int64(len(desired)) { return errors.New("one or more listener ids do not exist") }; if err := tx.Where("proxy_user_id = ? AND listener_id NOT IN ?", userID, desired).Delete(&models.ListenerUser{}).Error; err != nil { return err } } else if err := tx.Where("proxy_user_id = ?", userID).Delete(&models.ListenerUser{}).Error; err != nil { return err }; for _, listenerID := range desired { var binding models.ListenerUser; result := tx.Unscoped().Where("listener_id = ? AND proxy_user_id = ?", listenerID, userID).Find(&binding); if result.Error != nil { return result.Error }; if result.RowsAffected > 0 { if binding.DeletedAt.Valid { if err := tx.Unscoped().Model(&binding).Update("deleted_at", nil).Error; err != nil { return err } }; continue }; if err := tx.Create(&models.ListenerUser{ListenerID: listenerID, ProxyUserID: userID}).Error; err != nil { return err } }; return nil }) }
func (s *Service) GetListeners(userID uint) ([]models.Listener, error) { var listeners []models.Listener; err := s.db.Model(&models.Listener{}).Joins("JOIN listener_users ON listener_users.listener_id = listeners.id AND listener_users.deleted_at IS NULL").Where("listener_users.proxy_user_id = ?", userID).Order("listeners.id").Find(&listeners).Error; return listeners, err }
func IsCredentialActive(u models.ProxyUser) bool { now := time.Now(); if !u.Enabled { return false }; if !u.ExpireTime.IsZero() && !u.ExpireTime.After(now) { return false }; if u.TrafficLimit > 0 && u.TrafficUsed >= u.TrafficLimit { return false }; return true }

// ActiveCredentialsByListener reads credentials from the unified node config.
// For old databases, ValidateNode also backfills a stable generated credential
// before the export path is evaluated, so subscriptions no longer depend on
// Proxy User management.
func (s *Service) ActiveCredentialsByListener() (map[uint][]Credential, error) {
	result := make(map[uint][]Credential)
	var listeners []models.Listener
	if err := s.db.Where("enabled = ?", true).Find(&listeners).Error; err != nil { return nil, err }
	for _, listener := range listeners {
		before := listener.Config
		if err := node.ValidateNode(&listener); err != nil { return nil, fmt.Errorf("validate listener %q: %w", listener.Name, err) }
		if listener.Config != before {
			if err := s.db.Model(&models.Listener{}).Where("id = ?", listener.ID).Update("config", listener.Config).Error; err != nil { return nil, fmt.Errorf("save generated credentials for listener %q: %w", listener.Name, err) }
		}
		creds := credentialsFromListenerConfig(listener.Protocol, listener.Config)
		if len(creds) > 0 { result[listener.ID] = creds }
	}

	// Legacy fallback remains only for databases that already have ProxyUser
	// bindings and whose listener protocol has no embedded credential format.
	var rows []struct { ListenerID uint; ProxyUserID uint }
	if err := s.db.Model(&models.ListenerUser{}).Select("listener_id, proxy_user_id").Where("deleted_at IS NULL").Find(&rows).Error; err != nil { return nil, err }
	for _, row := range rows {
		if len(result[row.ListenerID]) > 0 { continue }
		u, err := s.GetByID(row.ProxyUserID); if err != nil { if errors.Is(err, gorm.ErrRecordNotFound) { continue }; return nil, err }
		if !IsCredentialActive(*u) { continue }
		password, err := decryptPassword(u.PasswordEncrypted); if err != nil { return nil, fmt.Errorf("decrypt proxy user %d password: %w", u.ID, err) }
		result[row.ListenerID] = append(result[row.ListenerID], Credential{Username: u.Username, Password: password, UUID: u.UUID})
	}
	return result, nil
}

func credentialsFromListenerConfig(protocol, raw string) []Credential {
	var cfg map[string]interface{}; if strings.TrimSpace(raw) == "" || json.Unmarshal([]byte(raw), &cfg) != nil { return nil }
	users, ok := cfg["users"]; if !ok { return nil }; result := []Credential{}
	switch protocol {
	case "anytls", "hysteria2", "mieru", "tuic":
		if m, ok := users.(map[string]interface{}); ok { for username, value := range m { result = append(result, Credential{Username: username, Password: fmt.Sprint(value), UUID: username}) } }
	default:
		if list, ok := users.([]interface{}); ok { for _, value := range list { row, ok := value.(map[string]interface{}); if !ok { continue }; result = append(result, Credential{Username: fmt.Sprint(row["username"]), Password: fmt.Sprint(row["password"]), UUID: fmt.Sprint(row["uuid"])}) } }
	}
	return result
}
func safeMask(s string) string { if len(s) <= 8 { return "********" }; return s[:4] + "..." + s[len(s)-4:] }
type SafeUser struct { ID uint `json:"id"`; Username string `json:"username"`; UUIDMasked string `json:"uuid_masked"`; TrafficLimit int64 `json:"traffic_limit"`; TrafficUsed int64 `json:"traffic_used"`; UploadBytes int64 `json:"upload_bytes"`; DownloadBytes int64 `json:"download_bytes"`; LastSeen *time.Time `json:"last_seen"`; Online bool `json:"online"`; ExpireTime time.Time `json:"expire_time"`; Enabled bool `json:"enabled"`; Blocked bool `json:"blocked"` }
func ToSafeUser(u *models.ProxyUser) SafeUser { return SafeUser{ID: u.ID, Username: u.Username, UUIDMasked: safeMask(u.UUID), TrafficLimit: u.TrafficLimit, TrafficUsed: u.TrafficUsed, UploadBytes: u.UploadBytes, DownloadBytes: u.DownloadBytes, LastSeen: u.LastSeen, Online: u.Online, ExpireTime: u.ExpireTime, Enabled: u.Enabled, Blocked: !IsCredentialActive(*u)} }
func encryptPassword(plain string) (string, error) { return security.Encrypt(plain) }
func decryptPassword(encoded string) (string, error) { return security.Decrypt(encoded) }
func newUUID() (string, error) { b := make([]byte, 16); if _, err := rand.Read(b); err != nil { return "", err }; b[6] = (b[6] & 0x0f) | 0x40; b[8] = (b[8] & 0x3f) | 0x80; return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil }
func randomToken(n int) (string, error) { b := make([]byte, n); if _, err := rand.Read(b); err != nil { return "", err }; return base64.RawURLEncoding.EncodeToString(b), nil }
