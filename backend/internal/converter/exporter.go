package converter

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/dzx941/3m-ui/backend/internal/config"
	"github.com/dzx941/3m-ui/backend/internal/database/models"
	"github.com/dzx941/3m-ui/backend/internal/user"
	"gorm.io/gorm"
)

// GenerateListenerURIs converts the same client proxy representation used by
// subscription generation into share URIs, keeping both export paths on one
// listener/credential source of truth.
func GenerateListenerURIs(db *gorm.DB, listener models.Listener, req *http.Request) ([]string, error) {
	if db == nil { return nil, fmt.Errorf("database is not initialized") }
	if !listener.Enabled { return nil, fmt.Errorf("listener is disabled") }
	server := ResolveServerAddress(config.GlobalConfig, req)
	credentials := []user.Credential{}
	if user.GlobalService != nil {
		byListener, err := user.GlobalService.ActiveCredentialsByListener()
		if err != nil { return nil, fmt.Errorf("failed to load listener credentials: %w", err) }
		credentials = byListener[listener.ID]
	}
	proxies, err := listenerToProxies(listener, server, credentials)
	if err != nil { return nil, err }
	uris := make([]string, 0, len(proxies))
	for _, proxy := range proxies {
		uri, err := proxyToURI(proxy)
		if err != nil { return nil, fmt.Errorf("listener %q: %w", listener.Name, err) }
		uris = append(uris, uri)
	}
	if len(uris) == 0 { return nil, fmt.Errorf("listener %q has no exportable client credentials", listener.Name) }
	return uris, nil
}

func proxyToURI(p map[string]interface{}) (string, error) {
	protocol := strings.ToLower(stringValue(p["type"]))
	server := stringValue(p["server"])
	port, ok := numericPort(p["port"])
	if protocol == "" || server == "" { return "", fmt.Errorf("proxy is missing type or server") }
	if !ok || port < 1 || port > 65535 { return "", fmt.Errorf("proxy has invalid port") }
	fragment := url.QueryEscape(stringValue(p["name"]))
	switch protocol {
	case "vless": return vlessURI(p, server, port, fragment)
	case "vmess": return vmessURI(p, server, port, stringValue(p["name"]))
	case "trojan": return passwordURI("trojan", p, server, port, fragment)
	case "hysteria2": return hysteria2URI(p, server, port, fragment)
	case "tuic": return tuicURI(p, server, port, fragment)
	case "anytls": return anyTLSURI(p, server, port, fragment)
	case "shadowsocks": return shadowsocksURI(p, server, port, fragment)
	default: return "", fmt.Errorf("URI export is not supported for protocol %q", protocol)
	}
}

func vlessURI(p map[string]interface{}, server string, port int, fragment string) (string, error) {
	uuid := stringValue(p["uuid"])
	if uuid == "" { return "", fmt.Errorf("VLESS proxy is missing uuid") }
	q := url.Values{}
	if v := stringValue(p["encryption"]); v != "" { q.Set("encryption", v) }
	if v := stringValue(p["flow"]); v != "" { q.Set("flow", v) }
	if boolValue(p["tls"]) { q.Set("security", "tls") }
	if v := firstString(p, "servername", "sni"); v != "" { q.Set("sni", v) }
	if v := stringValue(p["client-fingerprint"]); v != "" { q.Set("fp", v) }
	if network := stringValue(p["network"]); network != "" && network != "tcp" { q.Set("type", network) }
	addTransportURIParams(q, p)
	if reality, ok := p["reality-opts"].(map[string]interface{}); ok {
		q.Set("security", "reality")
		if v := stringValue(reality["public-key"]); v != "" { q.Set("pbk", v) }
		if v := firstValueString(reality, "short-id"); v != "" { q.Set("sid", v) }
	}
	if hasECH(p) && q.Get("security") == "" { q.Set("security", "tls") }
	return "vless://" + url.PathEscape(uuid) + "@" + formatHostPort(server, port) + queryAndFragment(q, fragment), nil
}

func vmessURI(p map[string]interface{}, server string, port int, name string) (string, error) {
	uuid := stringValue(p["uuid"])
	if uuid == "" { return "", fmt.Errorf("VMess proxy is missing uuid") }
	obj := map[string]string{"v":"2", "ps":name, "add":server, "port":strconv.Itoa(port), "id":uuid, "aid":strconv.Itoa(intValue(p["alterId"])), "scy":stringOr(p["cipher"], "auto")}
	if v := stringValue(p["network"]); v != "" && v != "tcp" { obj["net"] = v }
	if boolValue(p["tls"]) { obj["tls"] = "tls" }
	if v := firstString(p, "servername", "sni"); v != "" { obj["sni"] = v }
	if v := stringValue(p["client-fingerprint"]); v != "" { obj["fp"] = v }
	if reality, ok := p["reality-opts"].(map[string]interface{}); ok {
		obj["tls"] = "tls"
		if v := stringValue(reality["public-key"]); v != "" { obj["pbk"] = v }
		if v := firstValueString(reality, "short-id"); v != "" { obj["sid"] = v }
	}
	if ws, ok := p["ws-opts"].(map[string]interface{}); ok {
		obj["path"] = stringValue(ws["path"])
		if headers, ok := ws["headers"].(map[string]interface{}); ok { obj["host"] = stringValue(headers["Host"]) }
	}
	data, err := json.Marshal(obj)
	if err != nil { return "", err }
	return "vmess://" + base64.RawStdEncoding.EncodeToString(data), nil
}

func passwordURI(scheme string, p map[string]interface{}, server string, port int, fragment string) (string, error) {
	password := stringValue(p["password"])
	if password == "" { return "", fmt.Errorf("%s proxy is missing password", scheme) }
	q := url.Values{}
	if v := firstString(p, "sni", "servername"); v != "" { q.Set("sni", v) }
	if v := stringValue(p["client-fingerprint"]); v != "" { q.Set("fp", v) }
	if boolValue(p["skip-cert-verify"]) { q.Set("allowInsecure", "1") }
	if v := stringValue(p["network"]); v != "" && v != "tcp" { q.Set("type", v) }
	addTransportURIParams(q, p)
	if reality, ok := p["reality-opts"].(map[string]interface{}); ok {
		q.Set("security", "reality")
		if v := stringValue(reality["public-key"]); v != "" { q.Set("pbk", v) }
		if v := firstValueString(reality, "short-id"); v != "" { q.Set("sid", v) }
	}
	return scheme + "://" + url.PathEscape(password) + "@" + formatHostPort(server, port) + queryAndFragment(q, fragment), nil
}

func hysteria2URI(p map[string]interface{}, server string, port int, fragment string) (string, error) {
	password := stringValue(p["password"])
	if password == "" { return "", fmt.Errorf("Hysteria 2 proxy is missing password") }
	q := url.Values{}
	if v := firstString(p, "sni", "servername"); v != "" { q.Set("sni", v) }
	if boolValue(p["skip-cert-verify"]) { q.Set("insecure", "1") }
	if v := stringValue(p["obfs"]); v != "" { q.Set("obfs", v) }
	if v := stringValue(p["obfs-password"]); v != "" { q.Set("obfs-password", v) }
	if v := stringValue(p["up"]); v != "" { q.Set("up", v) }
	if v := stringValue(p["down"]); v != "" { q.Set("down", v) }
	if v := stringValue(p["fingerprint"]); v != "" { q.Set("pinSHA256", v) }
	return "hysteria2://" + url.PathEscape(password) + "@" + formatHostPort(server, port) + queryAndFragment(q, fragment), nil
}

func tuicURI(p map[string]interface{}, server string, port int, fragment string) (string, error) {
	q := url.Values{}
	if v := stringValue(p["congestion-controller"]); v != "" { q.Set("congestion_control", v) }
	if v := stringValue(p["bbr-profile"]); v != "" { q.Set("bbr_profile", v) }
	if v := stringValue(p["udp-relay-mode"]); v != "" { q.Set("udp_relay_mode", v) }
	if v := stringValue(p["alpn"]); v != "" { q.Set("alpn", v) }
	if boolValue(p["skip-cert-verify"]) { q.Set("allow_insecure", "1") }
	if v := firstString(p, "sni", "servername"); v != "" { q.Set("sni", v) }
	if token := stringValue(p["token"]); token != "" { return "tuic://" + url.PathEscape(token) + "@" + formatHostPort(server, port) + queryAndFragment(q, fragment), nil }
	uuid, password := stringValue(p["uuid"]), stringValue(p["password"])
	if uuid == "" || password == "" { return "", fmt.Errorf("TUIC proxy requires uuid/password or token") }
	return "tuic://" + url.PathEscape(uuid) + ":" + url.PathEscape(password) + "@" + formatHostPort(server, port) + queryAndFragment(q, fragment), nil
}

func anyTLSURI(p map[string]interface{}, server string, port int, fragment string) (string, error) {
	password := stringValue(p["password"])
	if password == "" { return "", fmt.Errorf("AnyTLS proxy is missing password") }
	q := url.Values{}
	if v := firstString(p, "sni", "servername"); v != "" { q.Set("sni", v) }
	if v := stringValue(p["client-fingerprint"]); v != "" { q.Set("fp", v) }
	if boolValue(p["skip-cert-verify"]) { q.Set("insecure", "1") }
	if v := stringValue(p["idle-session-check-interval"]); v != "" { q.Set("idle_session_check_interval", v) }
	if v := stringValue(p["idle-session-timeout"]); v != "" { q.Set("idle_session_timeout", v) }
	if v := stringValue(p["min-idle-session"]); v != "" { q.Set("min_idle_session", v) }
	return "anytls://" + url.PathEscape(password) + "@" + formatHostPort(server, port) + queryAndFragment(q, fragment), nil
}

func shadowsocksURI(p map[string]interface{}, server string, port int, fragment string) (string, error) {
	method, password := stringValue(p["cipher"]), stringValue(p["password"])
	if method == "" || password == "" { return "", fmt.Errorf("Shadowsocks proxy requires cipher/password") }
	userInfo := base64.RawStdEncoding.EncodeToString([]byte(method + ":" + password))
	return "ss://" + userInfo + "@" + formatHostPort(server, port) + "#" + fragment, nil
}

func addTransportURIParams(q url.Values, p map[string]interface{}) {
	if ws, ok := p["ws-opts"].(map[string]interface{}); ok {
		if path := stringValue(ws["path"]); path != "" { q.Set("path", path) }
		if headers, ok := ws["headers"].(map[string]interface{}); ok { if host := stringValue(headers["Host"]); host != "" { q.Set("host", host) } }
	}
	if grpc, ok := p["grpc-opts"].(map[string]interface{}); ok { if name := stringValue(grpc["grpc-service-name"]); name != "" { q.Set("serviceName", name) } }
	if xhttp, ok := p["xhttp-opts"].(map[string]interface{}); ok { if path := stringValue(xhttp["path"]); path != "" { q.Set("path", path) } }
}

func queryAndFragment(q url.Values, fragment string) string {
	out := ""
	if encoded := q.Encode(); encoded != "" { out = "?" + encoded }
	if fragment != "" { out += "#" + fragment }
	return out
}

func formatHostPort(host string, port int) string {
	if strings.Contains(host, ":") && !strings.HasPrefix(host, "[") { return "[" + host + "]:" + strconv.Itoa(port) }
	return host + ":" + strconv.Itoa(port)
}

func numericPort(v interface{}) (int, bool) {
	switch value := v.(type) {
	case int: return value, true
	case int64: return int(value), true
	case float64: return int(value), value == float64(int(value))
	case json.Number: n, err := strconv.Atoi(value.String()); return n, err == nil
	case string: n, err := strconv.Atoi(value); return n, err == nil
	default: return 0, false
	}
}

func intValue(v interface{}) int { n, _ := numericPort(v); return n }
func stringValue(v interface{}) string { if v == nil { return "" }; return fmt.Sprint(v) }
func stringOr(v interface{}, fallback string) string { if s := stringValue(v); s != "" { return s }; return fallback }
func firstString(m map[string]interface{}, keys ...string) string { for _, key := range keys { if v := stringValue(m[key]); v != "" { return v } }; return "" }
func firstValueString(m map[string]interface{}, keys ...string) string { return firstString(m, keys...) }
func hasECH(p map[string]interface{}) bool { cfg, ok := p["ech-opts"].(map[string]interface{}); return ok && boolValue(cfg["enable"]) }
