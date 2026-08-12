package node

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dzx941/3m-ui/backend/internal/database/models"
	mihomoConfig "github.com/dzx941/3m-ui/backend/internal/mihomo/config"
)

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
	l.Protocol = proto
	l.Type = proto

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

	var cfg map[string]interface{}
	if strings.TrimSpace(l.Config) != "" {
		if err := json.Unmarshal([]byte(l.Config), &cfg); err != nil {
			return fmt.Errorf("invalid listener configuration (must be valid JSON): %w", err)
		}
	}
	if cfg == nil {
		cfg = map[string]interface{}{}
	}

	if err := validateProtocolSpecific(proto, cfg); err != nil {
		return err
	}
	return nil
}

func validateProtocolSpecific(proto string, cfg map[string]interface{}) error {
	switch proto {
	case "snell":
		if value, ok := cfg["version"]; ok {
			if v, ok := value.(float64); ok && (v < 1 || v > 5) {
				return fmt.Errorf("snell version must be between 1 and 5")
			}
		}
	case "hysteria2":
		if obfs, ok := cfg["obfs"].(string); ok && obfs != "" && obfs != "salamander" {
			return fmt.Errorf("hysteria2 obfs must be salamander")
		}
		if err := requireCertificatePairOrAlternative(proto, cfg); err != nil {
			return err
		}
	case "anytls":
		if err := requireCertificatePairOrAlternative(proto, cfg); err != nil {
			return err
		}
	case "trusttunnel":
		if !hasCertificatePair(cfg) {
			return fmt.Errorf("trusttunnel listener requires certificate and private-key")
		}
	case "tuic":
		if _, hasToken := cfg["token"]; !hasToken {
			// V5 users may come from the user manager, so no token is required.
		}
	case "vless", "trojan":
		if err := requireCertificatePairOrAlternative(proto, cfg); err != nil {
			return err
		}
	case "sudoku":
		if min, ok := numeric(cfg["padding-min"]); ok {
			if max, ok := numeric(cfg["padding-max"]); ok && max < min {
				return fmt.Errorf("sudoku padding-max must be greater than or equal to padding-min")
			}
		}
	}
	return nil
}

func hasCertificatePair(cfg map[string]interface{}) bool {
	cert, certOK := cfg["certificate"].(string)
	key, keyOK := cfg["private-key"].(string)
	if !keyOK {
		key, keyOK = cfg["private_key"].(string)
	}
	return certOK && strings.TrimSpace(cert) != "" && keyOK && strings.TrimSpace(key) != ""
}

func requireCertificatePairOrAlternative(proto string, cfg map[string]interface{}) error {
	if hasCertificatePair(cfg) || boolValue(cfg["allow-insecure"]) {
		return nil
	}
	for _, key := range []string{"shadow-tls", "res-tls", "jls-config", "reality-config"} {
		if _, ok := cfg[key]; ok {
			return nil
		}
	}
	if proto == "vless" {
		if value, ok := cfg["decryption"]; ok && strings.TrimSpace(fmt.Sprint(value)) != "" {
			return nil
		}
	}
	return fmt.Errorf("%s listener requires certificate/private-key or a supported TLS alternative", proto)
}

func boolValue(v interface{}) bool {
	b, _ := v.(bool)
	return b
}

func numeric(v interface{}) (float64, bool) {
	n, ok := v.(float64)
	return n, ok
}
