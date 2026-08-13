package node

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/dzx941/3m-ui/backend/internal/database/models"
	mihomoConfig "github.com/dzx941/3m-ui/backend/internal/mihomo/config"
)

// listenerFieldSchema is the backend source of truth for fields that may be
// persisted in Listener.Config. The UI is not trusted: requests are validated
// against this schema before they reach Mihomo.
var listenerFieldSchema = map[string]map[string]struct{}{
	"shadowsocks": fields("cipher", "password", "udp", "shadow-tls", "res-tls", "jls-config", "kcp-tun", "mux-option"),
	"snell": fields("psk", "version", "udp", "obfs-opts", "shadow-tls", "res-tls", "jls-config"),
	"vmess": fields("users", "ws-path", "grpc-service-name", "mekya-config", "mkcp-config", "jls-config", "shadow-tls", "res-tls", "reality-config", "tlsmirror-config", "certificate", "private-key", "client-auth-type", "client-auth-cert", "ech-key", "allow-insecure", "mux-option"),
	"vless": fields("users", "ws-path", "grpc-service-name", "xhttp-config", "decryption", "reality-config", "shadow-tls", "res-tls", "jls-config", "certificate", "private-key", "client-auth-type", "client-auth-cert", "ech-key", "allow-insecure", "mux-option"),
	"trojan": fields("users", "ws-path", "grpc-service-name", "reality-config", "shadow-tls", "res-tls", "jls-config", "ss-option", "certificate", "private-key", "client-auth-type", "client-auth-cert", "ech-key", "allow-insecure", "mux-option"),
	"hysteria2": fields("users", "up", "down", "ignore-client-bandwidth", "obfs", "obfs-password", "masquerade", "bbr-profile", "alpn", "certificate", "private-key", "client-auth-type", "client-auth-cert", "ech-key", "allow-insecure", "mux-option"),
	"tuic": fields("users", "token", "certificate", "private-key", "client-auth-type", "client-auth-cert", "ech-key", "allow-insecure", "congestion-controller", "bbr-profile", "max-idle-time", "authentication-timeout", "alpn", "max-udp-relay-packet-size", "mux-option"),
	"shadowquic": fields("users", "jls-upstream", "alpn", "quic-versions", "zero-rtt", "congestion-controller", "up", "down", "ignore-client-bandwidth", "cwnd", "bbr-profile", "max-idle-time", "max-datagram-frame-size", "recv-window-conn", "recv-window", "disable-mtu-discovery"),
	"anytls": fields("users", "certificate", "private-key", "client-auth-type", "client-auth-cert", "ech-key", "allow-insecure", "padding-scheme"),
	"mieru": fields("transport", "users", "traffic-pattern", "user-hint-is-mandatory"),
	"sudoku": fields("key", "aead-method", "padding-min", "padding-max", "table-type", "custom-table", "custom-tables", "handshake-timeout", "enable-pure-downlink", "httpmask", "fallback", "mux-option"),
	"trusttunnel": fields("users", "certificate", "private-key", "client-auth-type", "client-auth-cert", "ech-key", "network", "congestion-controller", "bbr-profile"),
}

func fields(values ...string) map[string]struct{} {
	out := make(map[string]struct{}, len(values))
	for _, value := range values {
		out[value] = struct{}{}
	}
	return out
}

func ValidateNode(l *models.Listener) error {
	l.Name = strings.TrimSpace(l.Name)
	if l.Name == "" {
		return fmt.Errorf("listener name cannot be empty")
	}
	proto := strings.ToLower(strings.TrimSpace(l.Protocol))
	if proto == "" {
		proto = strings.ToLower(strings.TrimSpace(l.Type))
	}
	if !mihomoConfig.IsMihomoListenerProtocol(proto) {
		return fmt.Errorf("unsupported Mihomo listener protocol: %s", proto)
	}
	l.Protocol, l.Type = proto, proto
	if l.Port < 1 || l.Port > 65535 {
		return fmt.Errorf("port number must be between 1 and 65535")
	}
	l.BindAddress = strings.TrimSpace(l.BindAddress)
	if l.BindAddress == "" {
		l.BindAddress = strings.TrimSpace(l.Listen)
	}
	if l.BindAddress == "" {
		l.BindAddress = "0.0.0.0"
	}
	cfg, err := decodeConfig(l.Config)
	if err != nil {
		return err
	}
	if err := validateSchema(proto, cfg); err != nil {
		return err
	}
	return validateProtocolSpecific(proto, cfg)
}

func decodeConfig(raw string) (map[string]interface{}, error) {
	if strings.TrimSpace(raw) == "" {
		return map[string]interface{}{}, nil
	}
	var cfg map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return nil, fmt.Errorf("invalid listener configuration: %w", err)
	}
	if cfg == nil {
		return map[string]interface{}{}, nil
	}
	return cfg, nil
}

func validateSchema(proto string, cfg map[string]interface{}) error {
	allowed, ok := listenerFieldSchema[proto]
	if !ok {
		return fmt.Errorf("no listener schema registered for protocol %s", proto)
	}
	flat := make(map[string]interface{})
	flattenConfig("", cfg, flat)
	for key := range flat {
		top := strings.SplitN(key, ".", 2)[0]
		if _, ok := allowed[top]; !ok {
			return fmt.Errorf("%s listener: unsupported field %q", proto, key)
		}
	}
	return nil
}

func flattenConfig(prefix string, value interface{}, out map[string]interface{}) {
	if m, ok := value.(map[string]interface{}); ok {
		for key, child := range m {
			next := key
			if prefix != "" {
				next = prefix + "." + key
			}
			if childMap, ok := child.(map[string]interface{}); ok {
				flattenConfig(next, childMap, out)
			} else {
				out[next] = child
			}
		}
		return
	}
	if prefix != "" {
		out[prefix] = value
	}
}

func validateProtocolSpecific(proto string, cfg map[string]interface{}) error {
	if err := validateCertificateMode(proto, cfg); err != nil {
		return err
	}
	switch proto {
	case "snell":
		if value, ok := numeric(cfg["version"]); ok && (value < 1 || value > 5) {
			return fmt.Errorf("snell version must be between 1 and 5")
		}
	case "hysteria2":
		if obfs, ok := cfg["obfs"].(string); ok && obfs != "" && obfs != "salamander" {
			return fmt.Errorf("hysteria2 obfs must be salamander")
		}
		if !hasCertificatePair(cfg) && !boolValue(cfg["allow-insecure"]) {
			return fmt.Errorf("hysteria2 listener requires certificate/private-key or allow-insecure")
		}
		if users, ok := cfg["users"].(map[string]interface{}); ok && len(users) == 0 {
			return fmt.Errorf("hysteria2 users cannot be empty")
		}
	case "anytls":
		if _, ok := cfg["reality-config"]; ok {
			return fmt.Errorf("anytls does not support reality-config")
		}
		if !hasCertificatePair(cfg) && !boolValue(cfg["allow-insecure"]) {
			return fmt.Errorf("anytls listener requires certificate/private-key or allow-insecure")
		}
		if users, ok := cfg["users"].(map[string]interface{}); ok && len(users) == 0 {
			return fmt.Errorf("anytls users cannot be empty")
		}
	case "trusttunnel":
		if !hasCertificatePair(cfg) {
			return fmt.Errorf("trusttunnel listener requires certificate and private-key")
		}
	case "tuic":
		users := hasNonEmpty(cfg["users"])
		token := hasNonEmpty(cfg["token"])
		if users == token {
			return fmt.Errorf("tuic listener must configure exactly one of users (TUIC V5) or token (TUIC V4)")
		}
	case "vless", "vmess":
		if users, ok := cfg["users"].([]interface{}); ok {
			for i, user := range users {
				if err := validateUserRow(proto, i, user, true); err != nil {
					return err
				}
			}
		}
	case "trojan", "shadowquic":
		if users, ok := cfg["users"].([]interface{}); ok {
			for i, user := range users {
				if err := validateUserRow(proto, i, user, false); err != nil {
					return err
				}
			}
		}
	case "mieru":
		if users, ok := cfg["users"].(map[string]interface{}); ok && len(users) == 0 {
			return fmt.Errorf("mieru users cannot be empty")
		}
	case "sudoku":
		min, minOK := numeric(cfg["padding-min"])
		max, maxOK := numeric(cfg["padding-max"])
		if minOK && maxOK && max < min {
			return fmt.Errorf("sudoku padding-max must be greater than or equal to padding-min")
		}
	}
	return nil
}

func validateCertificateMode(proto string, cfg map[string]interface{}) error {
	cert := hasString(cfg["certificate"])
	key := hasString(cfg["private-key"]) || hasString(cfg["private_key"])
	if cert != key {
		return fmt.Errorf("%s listener requires certificate and private-key together", proto)
	}
	alternatives := 0
	if cert && key {
		alternatives++
	}
	for _, name := range []string{"reality-config", "shadow-tls", "res-tls", "jls-config"} {
		if hasNonEmpty(cfg[name]) {
			alternatives++
		}
	}
	if boolValue(cfg["allow-insecure"]) {
		alternatives++
	}
	if alternatives > 1 {
		return fmt.Errorf("%s listener has mutually exclusive TLS modes configured", proto)
	}
	if proto == "anytls" || proto == "hysteria2" || proto == "tuic" || proto == "trusttunnel" {
		if hasNonEmpty(cfg["reality-config"]) || hasNonEmpty(cfg["shadow-tls"]) || hasNonEmpty(cfg["res-tls"]) || hasNonEmpty(cfg["jls-config"]) {
			return fmt.Errorf("%s listener does not support the selected TLS alternative", proto)
		}
	}
	return nil
}

func validateUserRow(proto string, index int, raw interface{}, uuidMode bool) error {
	row, ok := raw.(map[string]interface{})
	if !ok {
		return fmt.Errorf("%s listener users[%d] must be an object", proto, index)
	}
	if !hasString(row["username"]) {
		return fmt.Errorf("%s listener users[%d] requires username", proto, index)
	}
	if uuidMode {
		if !hasString(row["uuid"]) {
			return fmt.Errorf("%s listener users[%d] requires uuid", proto, index)
		}
		if proto == "vmess" && hasNonEmpty(row["flow"]) {
			return fmt.Errorf("vmess listener users[%d] does not support flow", index)
		}
		if proto == "vless" && hasNonEmpty(row["alterId"]) {
			return fmt.Errorf("vless listener users[%d] does not support alterId", index)
		}
	} else if !hasString(row["password"]) {
		return fmt.Errorf("%s listener users[%d] requires password", proto, index)
	}
	return nil
}

func hasCertificatePair(cfg map[string]interface{}) bool {
	return hasString(cfg["certificate"]) && (hasString(cfg["private-key"]) || hasString(cfg["private_key"]))
}

func hasString(value interface{}) bool {
	s, ok := value.(string)
	return ok && strings.TrimSpace(s) != ""
}

func hasNonEmpty(value interface{}) bool {
	if value == nil {
		return false
	}
	if s, ok := value.(string); ok {
		return strings.TrimSpace(s) != ""
	}
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Slice, reflect.Map:
		return v.Len() > 0
	default:
		return true
	}
}

func boolValue(v interface{}) bool {
	b, _ := v.(bool)
	return b
}

func numeric(v interface{}) (float64, bool) {
	n, ok := v.(float64)
	return n, ok
}
