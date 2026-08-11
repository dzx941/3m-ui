package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dzx941/3m-ui/backend/internal/database/models"
	"gorm.io/gorm"
)

const DefaultTokenTTL = 24 * time.Hour

type LoginInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResult struct {
	Token              string    `json:"token"`
	ExpiresAt          time.Time `json:"expires_at"`
	Username           string    `json:"username"`
	Role               string    `json:"role"`
	MustChangePassword bool      `json:"must_change_password"`
}

func Login(db *gorm.DB, jwtSecret string, input LoginInput) (*LoginResult, error) {
	var user models.User
	if err := db.Where("username = ?", strings.TrimSpace(input.Username)).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid username or password")
		}
		return nil, err
	}
	if !CheckPasswordHash(input.Password, user.PasswordHash) {
		return nil, errors.New("invalid username or password")
	}
	token, exp, err := GenerateToken(jwtSecret, user.ID, user.Username, user.Role, DefaultTokenTTL)
	if err != nil {
		return nil, err
	}
	return &LoginResult{Token: token, ExpiresAt: exp, Username: user.Username, Role: user.Role, MustChangePassword: user.MustChangePassword}, nil
}

// EnsureAdmin creates the first administrator only when the database has no
// administrator. An explicit THREE_M_UI_ADMIN_PASSWORD is preferred.
// Otherwise a random password is written to the database directory so the
// installer/operator can retrieve it once.
func EnsureAdmin(db *gorm.DB, dbPath string) (created bool, username, password string, err error) {
	var count int64
	if err := db.Model(&models.User{}).Where("role = ?", "admin").Count(&count).Error; err != nil {
		return false, "", "", err
	}
	if count > 0 {
		return false, "", "", nil
	}

	username = strings.TrimSpace(os.Getenv("THREE_M_UI_ADMIN_USERNAME"))
	if username == "" {
		username = "admin"
	}
	password = os.Getenv("THREE_M_UI_ADMIN_PASSWORD")
	if password == "" {
		// Fresh installations use the simple default credentials requested by
		// the UI. The first successful login is forced to change the password.
		password = "admin"
	}

	hash, err := HashPassword(password)
	if err != nil {
		return false, "", "", err
	}
	if err := db.Create(&models.User{Username: username, PasswordHash: hash, Role: "admin", MustChangePassword: true}).Error; err != nil {
		return false, "", "", fmt.Errorf("create initial admin: %w", err)
	}

	return true, username, password, nil
}

func randomPassword(n int) (string, error) {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789"
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	for i := range buf {
		buf[i] = alphabet[int(buf[i])%len(alphabet)]
	}
	return string(buf), nil
}

func BearerToken(header string) string {
	const prefix = "Bearer "
	if strings.HasPrefix(header, prefix) {
		return strings.TrimSpace(strings.TrimPrefix(header, prefix))
	}
	return ""
}

func TokenFromRequest(authHeader string) string { return BearerToken(authHeader) }

func EncodePassword(password string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(password))
}
