package auth

import (
	"crypto/subtle"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/dzx941/3m-ui/backend/internal/config"
	"github.com/dzx941/3m-ui/backend/internal/database"
	"github.com/dzx941/3m-ui/backend/internal/database/models"
	"github.com/gin-gonic/gin"
)

type loginAttempt struct {
	count   int
	blocked time.Time
	last    time.Time
}

var loginLimiter = struct {
	sync.Mutex
	items map[string]loginAttempt
}{items: make(map[string]loginAttempt)}

const (
	loginWindow     = 15 * time.Minute
	loginMaxAttempt = 8
)

func allowLogin(ip string) bool {
	now := time.Now()
	loginLimiter.Lock()
	defer loginLimiter.Unlock()

	for key, attempt := range loginLimiter.items {
		if now.Sub(attempt.last) > loginWindow {
			delete(loginLimiter.items, key)
		}
	}

	attempt := loginLimiter.items[ip]
	if !attempt.blocked.IsZero() && now.Before(attempt.blocked) {
		return false
	}
	if attempt.last.IsZero() || now.Sub(attempt.last) > loginWindow {
		attempt.count = 0
	}
	attempt.last = now
	if attempt.count >= loginMaxAttempt {
		attempt.blocked = now.Add(loginWindow)
		loginLimiter.items[ip] = attempt
		return false
	}
	attempt.count++
	loginLimiter.items[ip] = attempt
	return true
}

func resetLoginLimit(ip string) {
	loginLimiter.Lock()
	delete(loginLimiter.items, ip)
	loginLimiter.Unlock()
}

func RegisterRoutes(rg *gin.RouterGroup, cfg *config.Config) {
	rg.POST("/login", func(c *gin.Context) {
		if !allowLogin(c.ClientIP()) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many login attempts; try again later"})
			return
		}

		var input LoginInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		result, err := Login(database.GlobalDB, cfg.JWT.Secret, input)
		if err != nil {
			status := http.StatusUnauthorized
			if err.Error() != "invalid username or password" {
				status = http.StatusInternalServerError
			}
			c.JSON(status, gin.H{"error": "invalid username or password"})
			return
		}
		resetLoginLimit(c.ClientIP())
		c.JSON(http.StatusOK, result)
	})

	rg.POST("/password", RequireAuth(cfg.JWT.Secret), func(c *gin.Context) {
		claims, ok := ClaimsFromContext(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}

		var req struct {
			CurrentPassword string `json:"current_password" binding:"required"`
			NewPassword     string `json:"new_password" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "current_password and new_password are required"})
			return
		}
		if len([]rune(req.NewPassword)) < 8 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "new password must be at least 8 characters"})
			return
		}
		if subtle.ConstantTimeCompare([]byte(req.NewPassword), []byte(req.CurrentPassword)) == 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "new password must differ from current password"})
			return
		}

		var user models.User
		if err := database.GlobalDB.First(&user, claims.UserID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		if !CheckPasswordHash(req.CurrentPassword, user.PasswordHash) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "current password is incorrect"})
			return
		}
		hash, err := HashPassword(req.NewPassword)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
			return
		}
		nextSessionVersion := user.SessionVersion + 1
		if nextSessionVersion == 0 {
			nextSessionVersion = 1
		}
		if err := database.GlobalDB.Model(&user).Updates(map[string]any{
			"password_hash":        hash,
			"must_change_password": false,
			"session_version":      nextSessionVersion,
		}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save password"})
			return
		}

		passwordFile := filepath.Join(filepath.Dir(cfg.Database.Path), ".initial_admin_password")
		if err := os.Remove(passwordFile); err != nil && !os.IsNotExist(err) {
			_ = err
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "password changed successfully"})
	})

	rg.GET("/me", RequireAuth(cfg.JWT.Secret), func(c *gin.Context) {
		claims, ok := ClaimsFromContext(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}
		var user models.User
		if err := database.GlobalDB.First(&user, claims.UserID).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"user_id":              claims.UserID,
			"username":             claims.Username,
			"role":                 claims.Role,
			"expires_at":           claims.ExpiresAt,
			"must_change_password": user.MustChangePassword,
		})
	})
}

func RequireAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := TokenFromRequest(c.GetHeader("Authorization"))
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}
		claims, err := ParseToken(secret, token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		var user models.User
		if err := database.GlobalDB.First(&user, claims.UserID).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			return
		}
		if !strings.EqualFold(user.Role, "admin") {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "administrator access required"})
			return
		}
		if user.SessionVersion == 0 || claims.SessionVersion == 0 || user.SessionVersion != claims.SessionVersion {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "session has been invalidated; please log in again"})
			return
		}

		c.Set("auth.claims", claims)
		c.Set("auth.user", &user)

		path := c.Request.URL.Path
		if path != "/api/v1/auth/password" && path != "/api/v1/auth/me" && user.MustChangePassword {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "password change required",
				"code":  "PASSWORD_CHANGE_REQUIRED",
			})
			return
		}

		c.Next()
	}
}

func ClaimsFromContext(c *gin.Context) (*JWTClaims, bool) {
	value, ok := c.Get("auth.claims")
	if !ok {
		return nil, false
	}
	claims, ok := value.(*JWTClaims)
	return claims, ok
}
