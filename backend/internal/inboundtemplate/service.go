package inboundtemplate

import (
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kazeyukiro/3m-ui/backend/internal/database/models"
	"github.com/kazeyukiro/3m-ui/backend/internal/user"
)

type Template struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Protocol    string   `json:"protocol"`
	Fields      []Field  `json:"fields"`
}

type Field struct {
	Name     string `json:"name"`
	Label    string `json:"label"`
	Required bool   `json:"required"`
	Secret   bool   `json:"secret"`
	Default  string `json:"default,omitempty"`
}

type CreateInput struct {
	TemplateID string            `json:"template_id" binding:"required"`
	Name       string            `json:"name"`
	Port       string            `json:"port"`
	Bind       string            `json:"bind"`
	Username   string            `json:"username"`
	Password   string            `json:"password"`
	UUID       string            `json:"uuid"`
	Values     map[string]string `json:"values"`
}

var templates = []Template{
	{ID: "vless-reality-tcp-vision", Name: "VLESS Reality TCP Vision", Description: "VLESS + Reality + XTLS Vision", Protocol: "vless", Fields: []Field{
		{Name: "dest", Label: "Reality 目标", Required: true, Default: "example.com:443"},
		{Name: "server_name", Label: "ServerName", Required: true, Default: "example.com"},
		{Name: "private_key", Label: "Reality PrivateKey", Required: false, Secret: true, Default: "auto"},
		{Name: "short_id", Label: "ShortID", Required: false, Secret: true, Default: "auto"},
	}},
	{ID: "vless-encryption-tcp-vision", Name: "VLESS Encryption TCP Vision", Description: "VLESS Encryption + XTLS Vision", Protocol: "vless", Fields: []Field{
		{Name: "decryption", Label: "Decryption", Required: true, Secret: true},
	}},
	{ID: "vless-enc-tls-xhttp-cdn", Name: "VLESS Encryption TLS XHTTP CDN", Description: "VLESS + Encryption + TLS + XHTTP", Protocol: "vless", Fields: []Field{
		{Name: "decryption", Label: "Decryption", Required: true, Secret: true},
		{Name: "certificate", Label: "证书路径", Required: true, Default: "./cert.pem"},
		{Name: "private_key", Label: "私钥路径", Required: true, Secret: true, Default: "./cert.key"},
		{Name: "host", Label: "XHTTP Host", Required: false},
		{Name: "path", Label: "XHTTP Path", Required: true, Default: "/xhttp"},
	}},
	{ID: "vless-tls-xhttp-h2-pinned", Name: "VLESS TLS XHTTP H2", Description: "VLESS + TLS + XHTTP", Protocol: "vless", Fields: []Field{
		{Name: "certificate", Label: "证书路径", Required: true, Default: "./server.pem"},
		{Name: "private_key", Label: "私钥路径", Required: true, Secret: true, Default: "./server.key"},
		{Name: "host", Label: "XHTTP Host", Required: false},
		{Name: "path", Label: "XHTTP Path", Required: true, Default: "/xhttp"},
	}},
	{ID: "hysteria2-ech", Name: "Hysteria2 ECH", Description: "Hysteria2 + TLS + ECH", Protocol: "hysteria2", Fields: []Field{
		{Name: "certificate", Label: "证书路径", Required: true, Default: "./server.crt"},
		{Name: "private_key", Label: "私钥路径", Required: true, Secret: true, Default: "./server.key"},
		{Name: "ech_key", Label: "ECH Key", Required: false, Secret: true},
		{Name: "up", Label: "上行限速", Required: false, Default: "50"},
		{Name: "down", Label: "下行限速", Required: false, Default: "20"},
	}},
	{ID: "shadowquic", Name: "ShadowQuic", Description: "ShadowQuic + JLS + QUIC", Protocol: "shadowquic", Fields: []Field{
		{Name: "jls_addr", Label: "JLS 目标", Required: true, Default: "example.com:443"},
		{Name: "jls_sni", Label: "JLS SNI", Required: true, Default: "example.com"},
	}},
}

func List() []Template { return append([]Template(nil), templates...) }

func Find(id string) (Template, bool) {
	for _, item := range templates {
		if item.ID == id {
			return item, true
		}
	}
	return Template{}, false
}

func Create(input CreateInput, nodeCreator func(*models.Listener) error, userCreator func(user.CreateInput) (*models.ProxyUser, error), binder func(uint, []uint) error) (*models.Listener, *models.ProxyUser, error) {
	t, ok := Find(strings.TrimSpace(input.TemplateID))
	if !ok {
		return nil, nil, fmt.Errorf("unknown inbound template %q", input.TemplateID)
	}
	port := strings.TrimSpace(input.Port)
	if port == "" {
		port = "443"
	}
	bind := strings.TrimSpace(input.Bind)
	if bind == "" {
		bind = "0.0.0.0"
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		name = t.Name
	}

	values := map[string]string{}
	for k, v := range input.Values { values[k] = strings.TrimSpace(v) }
	for _, f := range t.Fields {
		if values[f.Name] == "" && f.Default != "" && f.Default != "auto" { values[f.Name] = f.Default }
		if f.Required && values[f.Name] == "" { return nil, nil, fmt.Errorf("field %q is required", f.Label) }
	}

	cfg := map[string]interface{}{}
	var proxyUser *models.ProxyUser
	username := strings.TrimSpace(input.Username)
	password := input.Password
	uuid := strings.TrimSpace(input.UUID)
	if username == "" { username = "user1" }

	switch t.ID {
	case "vless-reality-tcp-vision":
		if uuid == "" { uuid = randomUUID() }
		privateKey := values["private_key"]
		if privateKey == "" || privateKey == "auto" { privateKey = realityPrivateKey() }
		shortID := values["short_id"]
		if shortID == "" || shortID == "auto" { shortID = randomHex(8) }
		cfg["users"] = []map[string]interface{}{{"username": username, "uuid": uuid, "flow": "xtls-rprx-vision"}}
		cfg["reality-config"] = map[string]interface{}{"dest": values["dest"], "private-key": privateKey, "short-id": []string{shortID}, "server-names": []string{values["server_name"]}}
	case "vless-encryption-tcp-vision":
		if uuid == "" { uuid = randomUUID() }
		cfg["users"] = []map[string]interface{}{{"username": username, "uuid": uuid, "flow": "xtls-rprx-vision"}}
		cfg["decryption"] = values["decryption"]
	case "vless-enc-tls-xhttp-cdn", "vless-tls-xhttp-h2-pinned":
		if uuid == "" { uuid = randomUUID() }
		cfg["users"] = []map[string]interface{}{{"username": username, "uuid": uuid}}
		if t.ID == "vless-enc-tls-xhttp-cdn" { cfg["decryption"] = values["decryption"] }
		cfg["certificate"], cfg["private-key"] = values["certificate"], values["private_key"]
		cfg["xhttp-config"] = map[string]interface{}{"host": values["host"], "path": values["path"], "mode": "auto"}
	case "hysteria2-ech":
		if password == "" { password = randomToken(24) }
		cfg["users"] = map[string]string{username: password}
		cfg["certificate"], cfg["private-key"] = values["certificate"], values["private_key"]
		if values["ech_key"] != "" { cfg["ech-key"] = values["ech_key"] }
		if values["up"] != "" { cfg["up"] = values["up"] }
		if values["down"] != "" { cfg["down"] = values["down"] }
		cfg["alpn"] = []string{"h3"}
	case "shadowquic":
		if password == "" { password = randomToken(24) }
		cfg["users"] = []map[string]string{{"username": username, "password": password}}
		cfg["quic-versions"] = []string{"v2"}
		cfg["congestion-controller"] = "bbr"
		cfg["jls-upstream"] = map[string]interface{}{"addr": values["jls_addr"], "sni": values["jls_sni"]}
		cfg["alpn"] = []string{"h3"}
	}

	encoded, err := json.Marshal(cfg)
	if err != nil { return nil, nil, fmt.Errorf("encode template: %w", err) }
	listener := &models.Listener{Name: name, Protocol: t.Protocol, Type: t.Protocol, Port: port, BindAddress: bind, Enabled: true, UDP: t.Protocol == "hysteria2" || t.Protocol == "shadowquic", Config: string(encoded)}
	if err := nodeCreator(listener); err != nil { return nil, nil, err }

	// Keep the generated credential in the normal 3m-ui user model so quota,
	// expiry, bindings and URI export remain consistent with manually-created users.
	if t.Protocol == "vless" || t.Protocol == "hysteria2" || t.Protocol == "shadowquic" {
		proxyUser, err = userCreator(user.CreateInput{Username: username, Password: password, UUID: uuid, Enabled: ptr(true)})
		if err != nil { return listener, nil, err }
		if err := binder(proxyUser.ID, []uint{listener.ID}); err != nil { return listener, proxyUser, err }
	}
	return listener, proxyUser, nil
}

func ptr(v bool) *bool { return &v }
func randomUUID() string { return randomHexUUID() }
func randomHexUUID() string { b := make([]byte, 16); _, _ = rand.Read(b); b[6] = (b[6]&0x0f)|0x40; b[8] = (b[8]&0x3f)|0x80; return fmt.Sprintf("%s-%s-%s-%s-%s", hex.EncodeToString(b[:4]), hex.EncodeToString(b[4:6]), hex.EncodeToString(b[6:8]), hex.EncodeToString(b[8:10]), hex.EncodeToString(b[10:])) }
func randomHex(n int) string { b := make([]byte, n); _, _ = rand.Read(b); return hex.EncodeToString(b) }
func randomToken(n int) string { b := make([]byte, n); _, _ = rand.Read(b); return base64.RawURLEncoding.EncodeToString(b) }
func realityPrivateKey() string { key, err := ecdh.X25519().GenerateKey(rand.Reader); if err != nil { return "" }; return base64.RawURLEncoding.EncodeToString(key.Bytes()) }
