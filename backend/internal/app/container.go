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

type Container struct {
	DB           *gorm.DB
	Config       *config.Config
	Mihomo       *mihomo.Service
	Listener     *listener.Service
	Node         *node.Service
	User         *user.Service
	System       *system.Service
	Traffic      *traffic.Service
	TrafficUser  *traffic.UserService
	Collector    *traffic.Collector
	Enforcer     *traffic.Enforcer
	Scheduler    *traffic.Scheduler
	ConfigEngine *dbconfig.ConfigEngine
}

func NewContainer(db *gorm.DB, cfg *config.Config) *Container {
	mihomoService := mihomo.NewService(cfg)
	userService := user.NewService(db)
	trafficService := traffic.NewService()
	trafficUserService := traffic.NewUserService(db)
	nodeService := node.NewService(db, cfg.Mihomo.Config)
	collector := traffic.NewCollectorFromDefaults(db, trafficService, trafficUserService)
	enforcer := traffic.NewEnforcer(db, nodeService)
	scheduler := traffic.NewScheduler(collector, enforcer, traffic.DefaultInterval)

	return &Container{
		DB:           db,
		Config:       cfg,
		Mihomo:       mihomoService,
		Listener:     listener.NewService(db, cfg.Mihomo.Config, mihomoService),
		Node:         nodeService,
		User:         userService,
		System:       system.NewService(),
		Traffic:      trafficService,
		TrafficUser:  trafficUserService,
		Collector:    collector,
		Enforcer:     enforcer,
		Scheduler:    scheduler,
		ConfigEngine: dbconfig.NewConfigEngine(db),
	}
}
