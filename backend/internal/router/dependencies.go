package router

import (
	"github.com/dzx941/3m-ui/backend/internal/config"
	"github.com/dzx941/3m-ui/backend/internal/listener"
	"github.com/dzx941/3m-ui/backend/internal/mihomo"
	mihomoConfig "github.com/dzx941/3m-ui/backend/internal/mihomo/config"
	"github.com/dzx941/3m-ui/backend/internal/node"
	"github.com/dzx941/3m-ui/backend/internal/system"
	"github.com/dzx941/3m-ui/backend/internal/traffic"
	"github.com/dzx941/3m-ui/backend/internal/user"
	"gorm.io/gorm"
)

// Dependencies is the router's explicit runtime dependency boundary.
//
// It deliberately lives in the router package instead of app to avoid an
// import cycle: app composes the router, while the router must not depend on
// the composition root. During the migration, legacy handlers may still use
// package globals; new handlers should consume this boundary instead.
type Dependencies struct {
	DB           *gorm.DB
	Config       *config.Config
	Mihomo       *mihomo.Service
	Listener     *listener.Service
	Node         *node.Service
	User         *user.Service
	System       *system.Service
	Traffic      *traffic.Service
	Collector    *traffic.Collector
	ConfigEngine *mihomoConfig.ConfigEngine
}

var runtimeDependencies Dependencies

// ConfigureDependencies installs the application-owned dependencies used by
// router handlers. It is called once by the application composition root.
func ConfigureDependencies(deps Dependencies) {
	runtimeDependencies = deps
}

// DependenciesSnapshot returns the currently configured router dependencies.
// It is primarily useful for handlers and tests while the remaining legacy
// global service references are migrated.
func DependenciesSnapshot() Dependencies {
	return runtimeDependencies
}
