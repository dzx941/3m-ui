package traffic

import (
	"log"
	"sync"

	"github.com/dzx941/3m-ui/backend/internal/database/models"
	"github.com/dzx941/3m-ui/backend/internal/node"
	"github.com/dzx941/3m-ui/backend/internal/user"
	"gorm.io/gorm"
)

// Enforcer watches for proxy users that cross into (or out of) a blocked
// state -- disabled, expired, or TrafficUsed >= TrafficLimit -- and, when
// that changes, triggers a Mihomo config regeneration so the exclusion
// takes effect promptly instead of waiting for the next manual node edit.
//
// The block/allow decision itself is never duplicated here: it is made
// exclusively by user.IsCredentialActive, the same predicate
// user.Service.ActiveCredentialsByListener uses when building the
// credentials that go into the generated Mihomo config. Enforcer only
// detects when that predicate's result changed for any user.
type Enforcer struct {
	db      *gorm.DB
	nodeSvc *node.Service

	mu          sync.Mutex
	lastBlocked map[uint]bool
}

// NewEnforcer builds an Enforcer. nodeSvc is used only to call
// RegenerateConfig() (which itself calls
// user.Service.ActiveCredentialsByListener() and attempts a Mihomo hot
// reload) -- no additional filtering logic lives here.
func NewEnforcer(db *gorm.DB, nodeSvc *node.Service) *Enforcer {
	return &Enforcer{db: db, nodeSvc: nodeSvc, lastBlocked: map[uint]bool{}}
}

// CheckAndEnforce recomputes the blocked set and regenerates + hot-reloads
// the Mihomo config only if it changed since the last check. Returns the
// number of currently blocked users for callers that want to report it
// (e.g. the dashboard), and any error encountered.
func (e *Enforcer) CheckAndEnforce() (blockedCount int, err error) {
	var users []models.ProxyUser
	if err := e.db.Find(&users).Error; err != nil {
		return 0, err
	}

	current := make(map[uint]bool, len(users))
	for _, u := range users {
		if !user.IsCredentialActive(u) {
			current[u.ID] = true
		}
	}

	e.mu.Lock()
	changed := !equalBlockedSets(e.lastBlocked, current)
	e.lastBlocked = current
	e.mu.Unlock()

	if !changed {
		return len(current), nil
	}

	log.Printf("traffic: enforcement state changed (%d user(s) now blocked); regenerating Mihomo config", len(current))
	if e.nodeSvc == nil {
		return len(current), nil
	}
	if err := e.nodeSvc.RegenerateConfig(); err != nil {
		return len(current), err
	}
	return len(current), nil
}

func equalBlockedSets(a, b map[uint]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if !b[k] {
			return false
		}
	}
	return true
}
