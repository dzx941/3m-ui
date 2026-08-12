package converter

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dzx941/3m-ui/backend/internal/config"
)

var SubconverterURL = "http://127.0.0.1:25500"

const maxSubconverterResponse = 8 << 20 // 8 MiB

var allowedTargets = map[string]struct{}{
	"clash":        {},
	"singbox":      {},
	"shadowrocket": {},
	"surge":        {},
}

// CallSubconverter calls the local subconverter service to convert standard configurations.
func CallSubconverter(cfg *config.Config, token string, target string, rawYAML []byte) ([]byte, error) {
	target = strings.ToLower(strings.TrimSpace(target))
	if _, ok := allowedTargets[target]; !ok {
		return nil, fmt.Errorf("unsupported conversion target: %s", target)
	}

	if strings.TrimSpace(token) == "" || strings.ContainsAny(token, "/?#") {
		return nil, fmt.Errorf("invalid access token")
	}

	port := 8080
	if cfg != nil && cfg.Server.Port != 0 {
		port = cfg.Server.Port
	}

	// Build raw source config URL pointing back to our own local endpoint.
	sourceURL := fmt.Sprintf("http://127.0.0.1:%d/api/v1/client/sub/%s?raw=true", port, url.PathEscape(token))

	u, err := url.Parse(strings.TrimRight(SubconverterURL, "/") + "/sub")
	if err != nil {
		return nil, fmt.Errorf("invalid subconverter URL: %w", err)
	}

	q := u.Query()
	q.Set("target", target)
	q.Set("url", sourceURL)
	u.RawQuery = q.Encode()

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("subconverter service unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
		return nil, fmt.Errorf("subconverter returned error status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	limited := io.LimitReader(resp.Body, maxSubconverterResponse+1)
	converted, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("failed to read subconverter response: %w", err)
	}
	if len(converted) > maxSubconverterResponse {
		return nil, fmt.Errorf("subconverter response exceeds %d bytes", maxSubconverterResponse)
	}
	if len(converted) == 0 {
		return nil, fmt.Errorf("subconverter returned an empty response")
	}

	_ = rawYAML // retained for API compatibility; subconverter fetches the authenticated raw endpoint.
	return converted, nil
}
