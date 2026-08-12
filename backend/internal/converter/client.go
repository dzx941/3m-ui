package converter

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/dzx941/3m-ui/backend/internal/config"
	"github.com/dzx941/3m-ui/backend/internal/database/models"
	mihomoConfig "github.com/dzx941/3m-ui/backend/internal/mihomo/config"
	"github.com/dzx941/3m-ui/backend/internal/security"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

func ResolveServerAddress(cfg *config.Config, req *http.Request) string {
	if envURL := os.Getenv("PUBLIC_URL"); envURL != "" { return cleanURLHost(envURL) }
	if cfg != nil && cfg.Server.PublicURL != "" { return cleanURLHost(cfg.Server.PublicURL) }
	if req != nil && req.Host != "" {
		host := req.Host
		if h, _, err := net.SplitHostPort(host); err == nil { return h }
		return strings.Trim(host, "[]")
	}
	return "127.0.0.1"
}

func cleanURLHost(raw string) string {
	raw = strings.TrimSpace(raw)
	u := raw
	if parsed, err := url.Parse(raw); err == nil && parsed.Host != "" { u = parsed.Host } else {
		u = strings.TrimPrefix(u, "https://")
		u = strings.TrimPrefix(u, "http://")
		u = strings.Split(u, "/")[0]
	}
	if h, _, err := net.SplitHostPort(u); err == nil { u = h }
	u = strings.Trim(u, "[]")
	if u == "" || strings.ContainsAny(u, "\r\n/\\") { return "127.0.0.1" }
	return u
}

func GetSubscriptionURL(cfg *config.Config, req *http.Request, token string, target string) string {
	var base string
	if envURL := os.Getenv("PUBLIC_URL"); envURL != "" { base = envURL
	} else if cfg != nil && cfg.Server.PublicURL != "" { base = cfg.Server.PublicURL
	} else if req != nil && req.Host != "" {
		scheme := "http"
		if req.TLS != nil { scheme = "https" }
		base = fmt.Sprintf("%s://%s", scheme, req.Host)
	} else { base = "http://127.0.0.1:8080" }
	base = strings.TrimSuffix(base, "/")
	pathToken := url.PathEscape(token)
	if target == "" { return fmt.Sprintf("%s/api/v1/client/sub/%s", base, pathToken) }
	return fmt.Sprintf("%s/api/v1/client/sub/%s?target=%s", base, pathToken, url.QueryEscape(strings.ToLower(strings.TrimSpace(target))))
}

func GenerateRawConfig(db *gorm.DB, token models.AccessToken, req *http.Request) ([]byte, error) {
	if db == nil { return nil, fmt.Errorf("database is not initialized") }
	if token.Type != "user" && token.Type != "proxy" { return nil, fmt.Errorf("invalid access token type") }

	proxies := make([]map[string]interface{}, 0)
	serverHost := ResolveServerAddress(config.GlobalConfig, req)

	if token.Type == "user" {
		var u models.ProxyUser
		if err := db.First(&u, token.TargetID).Error; err != nil {
			if err == gorm.ErrRecordNotFound { return nil, fmt.Errorf("proxy user not found") }
			return nil, fmt.Errorf("failed to fetch proxy user: %w", err)
		}
		if !u.Enabled || (!u.ExpireTime.IsZero() && u.ExpireTime.Before(timeNow())) { return nil, fmt.Errorf("proxy user is disabled or expired") }

		password := ""
		if u.PasswordEncrypted != "" {
			var err error
			password, err = security.Decrypt(u.PasswordEncrypted)
			if err != nil { return nil, fmt.Errorf("failed to decrypt proxy user credentials") }
		}

		var listeners []models.Listener
		err := db.Raw(`
			SELECT listeners.* FROM listeners
			JOIN listener_users ON listener_users.listener_id = listeners.id
			WHERE listener_users.proxy_user_id = ?
			AND listener_users.deleted_at IS NULL
			AND listeners.enabled = 1
		`, token.TargetID).Scan(&listeners).Error
		if err != nil { return nil, fmt.Errorf("failed to fetch user listeners: %w", err) }

		for _, l := range listeners {
			p, err := listenerToProxy(l, u.Username, password, u.UUID, serverHost)
			if err != nil { return nil, err }
			proxies = append(proxies, p)
		}
	} else {
		visual, err := mihomoConfig.GetVisualConfig(db)
		if err != nil { return nil, fmt.Errorf("failed to load visual config: %w", err) }
		idx := int(token.TargetID)
		if idx < 0 || idx >= len(visual.Proxies) { return nil, fmt.Errorf("proxy node not found") }
		pe := visual.Proxies[idx]
		p := map[string]interface{}{"name": pe.Name, "type": pe.Type, "server": pe.Server, "port": pe.Port}
		for k, v := range pe.Options { p[k] = v }
		proxies = append(proxies, p)
	}

	cfg := map[string]interface{}{
		"mixed-port": 7890,
		"proxies": proxies,
		"proxy-groups": []interface{}{map[string]interface{}{"name": "PROXY", "type": "select", "proxies": getProxyNames(proxies)}},
		"rules": []string{"MATCH,PROXY"},
	}
	return yaml.Marshal(cfg)
}

func listenerToProxy(l models.Listener, username, password, uuid, server string) (map[string]interface{}, error) {
	protocol := strings.ToLower(strings.TrimSpace(l.Protocol))
	if protocol == "" { protocol = strings.ToLower(strings.TrimSpace(l.Type)) }
	p := map[string]interface{}{"name": l.Name, "type": protocol, "server": server, "port": l.Port, "udp": l.UDP}

	opts := map[string]interface{}{}
	if l.Config != "" {
		if err := json.Unmarshal([]byte(l.Config), &opts); err != nil { return nil, fmt.Errorf("invalid listener config for %q: %w", l.Name, err) }
	}

	if l.TLS || hasAny(opts, "certificate", "private-key", "private_key") { p["tls"] = true }
	copyOption(p, opts, "servername", "servername")
	copyOption(p, opts, "sni", "sni")
	copyOption(p, opts, "skip-cert-verify", "skip-cert-verify")
	copyOption(p, opts, "name-cert-verify", "name-cert-verify")
	copyOption(p, opts, "fingerprint", "fingerprint")
	copyOption(p, opts, "client-fingerprint", "client-fingerprint")
	copyOption(p, opts, "alpn", "alpn")

	switch protocol {
	case "shadowsocks":
		copyOption(p, opts, "cipher", "cipher")
		if password != "" { p["password"] = password } else if v, ok := opts["password"]; ok { p["password"] = v }
		copyOption(p, opts, "plugin", "plugin")
		copyOption(p, opts, "plugin-opts", "plugin-opts")
	case "vmess":
		p["uuid"] = uuid
		copyOption(p, opts, "alterId", "alterId")
		copyOption(p, opts, "cipher", "cipher")
		copyOption(p, opts, "packet-encoding", "packet-encoding")
		copyOption(p, opts, "global-padding", "global-padding")
		copyOption(p, opts, "authenticated-length", "authenticated-length")
	case "vless":
		p["uuid"] = uuid
		copyOption(p, opts, "flow", "flow")
		copyOption(p, opts, "encryption", "encryption")
		copyOption(p, opts, "packet-encoding", "packet-encoding")
		if reality := realityOpts(opts); reality != nil { p["reality-opts"] = reality }
	case "trojan":
		p["password"] = password
		if reality := realityOpts(opts); reality != nil { p["reality-opts"] = reality }
		if shadow := shadowTLSOpts(opts); shadow != nil { p["shadow-tls-opts"] = shadow }
		if restls, ok := opts["restls-opts"]; ok { p["restls-opts"] = restls }
		if jls, ok := opts["jls-opts"]; ok { p["jls-opts"] = jls }
	case "hysteria2":
		p["password"] = password
		for _, key := range []string{"up", "down", "obfs", "obfs-password", "masquerade", "ignore-client-bandwidth", "bbr-profile"} { copyOption(p, opts, key, key) }
	case "tuic":
		p["uuid"] = uuid
		p["password"] = password
		for _, key := range []string{"congestion-controller", "udp-relay-mode", "max-idle-time", "reduce-rtt", "request-timeout", "heartbeat-interval", "alpn", "disable-sni", "max-udp-relay-packet-size"} { copyOption(p, opts, key, key) }
	default:
		return nil, fmt.Errorf("unsupported listener protocol %q for client export", protocol)
	}

	if network, ok := opts["network"].(string); ok && network != "" { p["network"] = network }
	if wsPath, ok := opts["ws-path"].(string); ok && wsPath != "" { p["network"] = "ws"; p["ws-opts"] = map[string]interface{}{"path": wsPath} }
	if grpcName, ok := opts["grpc-service-name"].(string); ok && grpcName != "" { p["network"] = "grpc"; p["grpc-opts"] = map[string]interface{}{"grpc-service-name": grpcName} }
	return p, nil
}

func copyOption(dst, src map[string]interface{}, srcKey, dstKey string) {
	if value, ok := src[srcKey]; ok { dst[dstKey] = value }
}

func hasAny(m map[string]interface{}, keys ...string) bool {
	for _, key := range keys { if _, ok := m[key]; ok { return true } }
	return false
}

func realityOpts(src map[string]interface{}) map[string]interface{} {
	if raw, ok := src["reality-opts"].(map[string]interface{}); ok { return raw }
	cfg, ok := src["reality-config"].(map[string]interface{})
	if !ok { return nil }
	result := map[string]interface{}{}
	if v, ok := cfg["public-key"]; ok { result["public-key"] = v }
	if v, ok := cfg["short-id"]; ok {
		switch ids := v.(type) {
		case []interface{}:
			if len(ids) > 0 { result["short-id"] = ids[0] }
		case string:
			result["short-id"] = ids
		}
	}
	if len(result) == 0 { return nil }
	return result
}

func shadowTLSOpts(src map[string]interface{}) map[string]interface{} {
	if raw, ok := src["shadow-tls-opts"].(map[string]interface{}); ok { return raw }
	cfg, ok := src["shadow-tls"].(map[string]interface{})
	if !ok { return nil }
	result := map[string]interface{}{}
	if v, ok := cfg["version"]; ok { result["version"] = v }
	if v, ok := cfg["password"]; ok { result["password"] = v }
	if len(result) == 0 { return nil }
	return result
}

func getProxyNames(proxies []map[string]interface{}) []string {
	names := make([]string, 0, len(proxies))
	for _, p := range proxies {
		if name, ok := p["name"].(string); ok && strings.TrimSpace(name) != "" { names = append(names, name) }
	}
	if len(names) == 0 { names = append(names, "DIRECT") }
	return names
}

var timeNow = func() time.Time { return time.Now() }
