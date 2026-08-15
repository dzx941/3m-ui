package router

import (
	"github.com/dzx941/3m-ui/backend/internal/config"
	"github.com/dzx941/3m-ui/backend/internal/mihomo"
	"github.com/dzx941/3m-ui/backend/internal/node"
	"github.com/dzx941/3m-ui/backend/internal/system"
	"github.com/dzx941/3m-ui/backend/internal/traffic"
	"github.com/dzx941/3m-ui/backend/internal/user"
	"gorm.io/gorm"
)

// Deps holds runtime dependencies for HTTP handlers.
// Production wiring comes from app.Container.RouterDeps().
type Deps struct {
	DB               *gorm.DB
	Config           *config.Config
	Mihomo           *mihomo.Service
	Traffic          *traffic.Service
	TrafficCollector *traffic.Collector
	User             *user.Service
	Node             *node.Service
	System           *system.Service
}

func (d Deps) mihomoService() *mihomo.Service       { return d.Mihomo }
func (d Deps) trafficService() *traffic.Service     { return d.Traffic }
func (d Deps) trafficCollector() *traffic.Collector { return d.TrafficCollector }
func (d Deps) userService() *user.Service           { return d.User }
func (d Deps) nodeService() *node.Service           { return d.Node }
func (d Deps) systemService() *system.Service       { return d.System }
