package auth

import (
	"net/http"

	"github.com/dzx941/3m-ui/backend/internal/config"
	"github.com/dzx941/3m-ui/backend/internal/database"
	"github.com/dzx941/3m-ui/backend/internal/database/models"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, cfg *config.Config) {
	rg.POST("/login", func(c *gin.Context) {
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
			c.JSON(status, gin.H{"error": err.Error()})
			return
		}
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
			NewPassword string `json:"new_password" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "current_password and new_password are required"})
			return
		}
		if len(req.NewPassword) < 8 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "new password must be at least 8 characters"})
			return
		}
		if req.NewPassword == req.CurrentPassword {
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
		if err := database.GlobalDB.Model(&user).Updates(map[string]any{
			"password_hash": hash,
			"must_change_password": false,
		}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save password"})
			return
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
			"user_id": claims.UserID,
			"username": claims.Username,
			"role": claims.Role,
			"expires_at": claims.ExpiresAt,
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
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.Set("auth.claims", claims)
		c.Next()
	}
}

// RequirePasswordChanged prevents access to management APIs until the
// initial password has been replaced. Login and /auth/password remain usable.
func RequirePasswordChanged() gin.HandlerFunc {
	return func(c *gin.Context) {
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
		if user.MustChangePassword {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "password change required", "code": "PASSWORD_CHANGE_REQUIRED"})
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
