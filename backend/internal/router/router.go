package router

import (
	"net/http"

	"github.com/dzx941/3m-ui/backend/internal/auth"
	"github.com/dzx941/3m-ui/backend/internal/config"
	"github.com/dzx941/3m-ui/backend/internal/node"
	"github.com/dzx941/3m-ui/backend/internal/system"
	"github.com/dzx941/3m-ui/backend/internal/traffic"
	"github.com/dzx941/3m-ui/backend/internal/user"
	"github.com/gin-gonic/gin"
)

// SetupRouter builds the Gin engine from config only (tests / legacy).
func SetupRouter(cfg *config.Config) *gin.Engine {
	return SetupRouterWithDeps(Deps{Config: cfg})
}

// SetupRouterWithDeps builds the Gin engine using explicit dependencies.
// Route paths and response JSON remain unchanged for API compatibility.
func SetupRouterWithDeps(d Deps) *gin.Engine {
	cfg := d.Config
	if cfg == nil {
		cfg = &config.Config{}
	}
	db := resolveDB(d)

	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.Default()
	r.Use(CORSMiddleware(cfg.Security.CORSOrigins))

	apiV1 := r.Group("/api/v1")
	{
		auth.NewHandler(db, cfg).RegisterRoutes(apiV1.Group("/auth"))

		apiV1.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		registerPublicClientRoutes(apiV1, d)

		apiV1.Use(auth.RequireAuth(db, cfg.JWT.Secret))

		registerAccessTokenRoutes(apiV1, d, cfg)
		registerDashboardRoute(apiV1, d)

		system.NewHandler(d.systemService()).RegisterRoutes(apiV1.Group("/system"))
		registerMihomoRoutes(apiV1, d)

		user.NewHandler(d.userService()).RegisterRoutes(apiV1.Group("/users"))

		nodeHandler := node.NewHandler(d.nodeService(), d.userService(), db, cfg)
		nodeHandler.RegisterRoutes(apiV1.Group("/nodes"))
		nodeHandler.RegisterRoutes(apiV1.Group("/listeners"))

		traffic.RegisterRoutes(
			apiV1.Group("/traffic"),
			traffic.NewHandler(d.trafficService(), d.trafficCollector(), db),
		)

		registerConfigRoutes(apiV1, d, cfg)
	}

	return r
}
