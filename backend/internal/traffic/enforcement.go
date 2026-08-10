package traffic

import (
	"log"
	"sync"

	"github.com/dzx941/3m-ui/backend/internal/mihomo"
	"github.com/dzx941/3m-ui/backend/internal/user"
	"gorm.io/gorm"
)

var (
	enforceMu            sync.Mutex
	lastActiveUsers      map[uint]bool
	RegenerateConfigFunc func() error
)

func EnforceLimits(db *gorm.DB, configPath string) {
	enforceMu.Lock()
	defer enforceMu.Unlock()

	if user.GlobalService == nil {
		return
	}

	// 1. Get current active credentials
	activeCreds, err := user.GlobalService.ActiveCredentialsByListener()
	if err != nil {
		log.Printf("Enforcement failed to get active credentials: %v", err)
		return
	}

	// 2. Extract current set of active user IDs
	currentActiveUsers := make(map[uint]bool)
	for _, credList := range activeCreds {
		for _, cred := range credList {
			currentActiveUsers[cred.ID] = true
		}
	}

	// 3. Compare with last active user set. If different, we must regenerate and reload!
	if lastActiveUsers == nil {
		// Initialize on first run
		lastActiveUsers = currentActiveUsers
		return
	}

	changed := false
	if len(currentActiveUsers) != len(lastActiveUsers) {
		changed = true
	} else {
		for id := range currentActiveUsers {
			if !lastActiveUsers[id] {
				changed = true
				break
			}
		}
	}

	if changed {
		log.Printf("Traffic limits or expiration changes detected. Regenerating and reloading Mihomo config...")
		lastActiveUsers = currentActiveUsers

		// A. Regenerate config
		if RegenerateConfigFunc != nil {
			if err := RegenerateConfigFunc(); err != nil {
				log.Printf("Enforcement failed to regenerate config: %v", err)
				return
			}
		} else {
			log.Printf("Enforcement warning: RegenerateConfigFunc callback not registered")
		}

		// B. Reload Mihomo config
		baseURL, secret := getMihomoAPIConfig(configPath)
		apiClient := mihomo.NewExternalControllerAPI(baseURL, secret)
		payload := map[string]any{
			"path": configPath,
		}
		if err := apiClient.ReloadConfig(payload); err != nil {
			log.Printf("Enforcement failed to reload config via API: %v. Attempting process restart...", err)
			// Fallback: Restart if Reload fails
			if restartErr := mihomo.GlobalService.RestartMihomo(); restartErr != nil {
				log.Printf("Enforcement fallback process restart failed: %v", restartErr)
			} else {
				log.Printf("Enforcement fallback process restart succeeded.")
			}
		} else {
			log.Printf("Enforcement successfully hot-reloaded config in Mihomo Core.")
		}
	}
}
