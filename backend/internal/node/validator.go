package node

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/dzx941/3m-ui/backend/internal/credentials"
	"github.com/dzx941/3m-ui/backend/internal/database/models"
	mihomoConfig "github.com/dzx941/3m-ui/backend/internal/mihomo/config"
)

// ValidateNode is the backend trust boundary for every Listener. The UI is
// never trusted: protocol names, common fields, nested objects and protocol
// specific invariants are checked before configuration reaches Mihomo.
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
	if err := credentials.EnsureListenerCredentials(l); err != nil {
		return err
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

// validateSchema rejects unknown top-level fields and, importantly, unknown
// nested keys. The old validator only checked the first path segment, which
// meant payloads such as {"xhttp-config":{"unknown":true}} could bypass the
// advertised strict schema.
func validateSchema(proto string, cfg map[string]interface{}) error {
	schema, ok := mihomoConfig.GetMihomoListenerSchema(proto)
	if !ok {
		return fmt.Errorf("no listener schema registered for protocol %s", proto)
	}
	return validateObject(proto, "", cfg, schema.Fields, schema.NestedFields)
}

func validateObject(proto, prefix string, value map[string]interface{}, allowed map[string]struct{}, nested map[string]map[string]struct{}) error {
	for key, child := range value {
		if _, ok := allowed[key]; !ok {
			field := key
			if prefix != "" {
				field = prefix + "." + key
			}
			return fmt.Errorf("%s listener: unsupported field %q", proto, field)
		}
		if childMap, ok := child.(map[string]interface{}); ok {
			childAllowed, hasNestedSchema := nested[key]
			if !hasNestedSchema {
				// A registered scalar field must not silently become an arbitrary
				// object. Complex fields must have an explicit nested schema.
				return fmt.Errorf("%s listener: field %q does not accept an object", proto, joinField(prefix, key))
			}
			childNested := make(map[string]map[string]struct{})
			for nestedKey, nestedAllowed := range nested {
				if strings.HasPrefix(nestedKey, key+".") {
					childNested[strings.TrimPrefix(nestedKey, key+".")] = nestedAllowed
				}
			}
			if err := validateObject(proto, joinField(prefix, key), childMap, childAllowed, childNested); err != nil {
				return err
			}
		}
	}
	return nil
}

func joinField(prefix, field string) string {
	if prefix == "" {
		return field
	}
	return prefix + "." + field
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
		users, ok := cfg["users"].([]interface{})
		if !ok || len(users) == 0 {
			return fmt.Errorf("%s listener requires at least one user", proto)
		}
		for i, user := range users {
			if err := validateUserRow(proto, i, user, true); err != nil {
				return err
			}
		}
	case "trojan", "shadowquic":
		users, ok := cfg["users"].([]interface{})
		if !ok || len(users) == 0 {
			return fmt.Errorf("%s listener requires at least one user", proto)
		}
		for i, user := range users {
			if err := validateUserRow(proto, i, user, false); err != nil {
				return err
			}
		}
	case "mieru":
		if users, ok := cfg["users"].(map[string]interface{}); !ok || len(users) == 0 {
			return fmt.Errorf("mieru listener requires at least one user")
		}
	case "sudoku":
		min, minOK := numeric(cfg["padding-min"])
		max, maxOK := numeric(cfg["padding-max"])
		if minOK && maxOK && max < min {
			return fmt.Errorf("sudoku padding-max must be greater than or equal to padding-min")
		}
		if cfg["enable-pure-downlink"] == false && cfg["aead-method"] == "none" {
			return fmt.Errorf("sudoku requires aead-method other than none when pure downlink is disabled")
		}
	}
	return nil
}

func validateCertificateMode(proto string, cfg map[string]interface{}) error {
	cert := hasString(cfg["certificate"])
	key := hasString(cfg["private-key"])
	if cert != key {
		return fmt.Errorf("%s listener requires certificate and private-key together", proto)
	}

	if proto == "vless" || proto == "vmess" || proto == "trojan" {
		if hasNonEmpty(cfg["reality-config"]) && (cert || key) {
			return fmt.Errorf("%s listener cannot combine reality-config with certificate/private-key", proto)
		}
	}
	if proto == "anytls" && hasNonEmpty(cfg["reality-config"]) {
		return fmt.Errorf("anytls listener does not support reality-config")
	}
	if hasNonEmpty(cfg["reality-config"]) && proto != "vless" && proto != "vmess" && proto != "trojan" {
		return fmt.Errorf("%s listener does not support reality-config", proto)
	}
	return nil
}

func validateUserRow(proto string, index int, raw interface{}, uuidMode bool) error {
	row, ok := raw.(map[string]interface{})
	if !ok {
		return fmt.Errorf("%s listener users[%d] must be an object", proto, index)
	}
	allowed := map[string]struct{}{"username": {}}
	if uuidMode {
		allowed["uuid"] = struct{}{}
		if proto == "vmess" {
			allowed["alterId"] = struct{}{}
		} else {
			allowed["flow"] = struct{}{}
		}
	} else {
		allowed["password"] = struct{}{}
	}
	for key := range row {
		if _, ok := allowed[key]; !ok {
			return fmt.Errorf("%s listener users[%d]: unsupported field %q", proto, index, key)
		}
	}
	if !hasString(row["username"]) {
		return fmt.Errorf("%s listener users[%d] requires username", proto, index)
	}
	if uuidMode {
		if !hasString(row["uuid"]) {
			return fmt.Errorf("%s listener users[%d] requires uuid", proto, index)
		}
	} else if !hasString(row["password"]) {
		return fmt.Errorf("%s listener users[%d] requires password", proto, index)
	}
	return nil
}

func hasCertificatePair(cfg map[string]interface{}) bool {
	return hasString(cfg["certificate"]) && hasString(cfg["private-key"])
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
