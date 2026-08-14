package app

import (
	"github.com/dzx941/3m-ui/backend/internal/config"
	"github.com/dzx941/3m-ui/backend/internal/database"
	"github.com/dzx941/3m-ui/backend/internal/listener"
	"github.com/dzx941/3m-ui/backend/internal/mihomo"
	dbconfig "github.com/dzx941/3m-ui/backend/internal/mihomo/config"
	"github.com/dzx941/3m-ui/backend/internal/node"
	"github.com/dzx941/3m-ui/backend/internal/system"
	"github.com/dzx941/3m-ui/backend/internal/traffic"
	"github.com/dzx941/3m-ui/backend/internal/user"
	"gorm.io/gorm"
)

// Container is the application composition root. It makes the core runtime
// dependencies explicit while legacy package globals remain temporarily
// populated for handlers that have not yet been migrated to dependency
// injection.
type Container struct {
	DB           *gorm.DB
	Config       *config.Config
	Mihomo       *mihomo.Service
	Listener     *listener.Service
	Node         *node.Service
	User         *user.Service
	System       *system.Service
	Traffic      *traffic.Service
	ConfigEngine *dbconfig.ConfigEngine
}

// NewContainer constructs the application services from one database and one
// configuration. Package globals are registered here, in one place, so the
// rest of the application no longer needs to know how services are bootstrapped.
func NewContainer(db *gorm.DB, cfg *config.Config) *Container {
	mihomo.InitService(cfg)
	listener.InitService(db, cfg.Mihomo.Config)
	node.InitService(db, cfg.Mihomo.Config)
	user.InitService(db)
	system.InitService()
	trafficService := traffic.InitGlobalService()

	return &Container{
		DB:           db,
		Config:       cfg,
		Mihomo:       mihomo.GlobalService,
		Listener:     listener.GlobalService,
		Node:         node.GlobalService,
		User:         user.GlobalService,
		System:       system.GlobalService,
		Traffic:      trafficService,
		ConfigEngine: dbconfig.NewConfigEngine(db),
	}
}
