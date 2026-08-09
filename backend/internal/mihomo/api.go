package mihomo

import (
	"fmt"
	"net/http"
	"time"
)

// ExternalControllerAPI wraps communication with the Mihomo External Controller REST API
type ExternalControllerAPI struct {
	BaseURL string
	Secret  string
	client  *http.Client
}

func NewExternalControllerAPI(baseURL, secret string) *ExternalControllerAPI {
	return &ExternalControllerAPI{
		BaseURL: baseURL,
		Secret:  secret,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// GetConfig fetches the current configuration from Mihomo's external API
func (api *ExternalControllerAPI) GetConfig() (string, error) {
	// For now, return placeholder or perform http request if service is running
	return "{}", nil
}

// ReloadConfig sends a POST request to trigger config reload in Mihomo
func (api *ExternalControllerAPI) ReloadConfig(payload map[string]interface{}) error {
	// For future implementation
	return fmt.Errorf("not implemented yet")
}
