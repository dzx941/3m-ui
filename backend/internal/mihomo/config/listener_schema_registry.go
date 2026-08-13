package config

// ListenerSchema describes the fields accepted in Listener.Config for one
// Mihomo listener protocol. It deliberately models only the protocol payload;
// listener-level fields (name, type, port, listen, rule and proxy) are emitted
// by the generator itself.
type ListenerSchema struct {
	Protocol string
	Fields   map[string]struct{}
}

func listenerFields(values ...string) map[string]struct{} {
	fields := make(map[string]struct{}, len(values))
	for _, value := range values {
		fields[value] = struct{}{}
	}
	return fields
}

// MihomoListenerSchemas is the single backend registry used by validation.
// The registry is intentionally kept separate from the HTTP/API layer so the
// same schema can be consumed by generators and tests without duplicating the
// protocol switch in multiple packages.
var MihomoListenerSchemas = map[string]ListenerSchema{
	"shadowsocks": {Protocol: "shadowsocks", Fields: listenerFields(
		"cipher", "password", "udp", "shadow-tls", "res-tls", "jls-config", "kcp-tun", "mux-option",
	)},
	"snell": {Protocol: "snell", Fields: listenerFields(
		"psk", "version", "udp", "obfs-opts", "shadow-tls", "res-tls", "jls-config",
	)},
	"vmess": {Protocol: "vmess", Fields: listenerFields(
		"users", "ws-path", "grpc-service-name", "mekya-config", "mkcp-config", "jls-config",
		"shadow-tls", "res-tls", "reality-config", "tlsmirror-config", "certificate", "private-key",
		"client-auth-type", "client-auth-cert", "ech-key", "allow-insecure", "mux-option",
	)},
	"vless": {Protocol: "vless", Fields: listenerFields(
		"users", "ws-path", "grpc-service-name", "xhttp-config", "decryption", "reality-config",
		"shadow-tls", "res-tls", "jls-config", "certificate", "private-key", "client-auth-type",
		"client-auth-cert", "ech-key", "allow-insecure", "mux-option",
	)},
	"trojan": {Protocol: "trojan", Fields: listenerFields(
		"users", "ws-path", "grpc-service-name", "reality-config", "shadow-tls", "res-tls", "jls-config",
		"ss-option", "certificate", "private-key", "client-auth-type", "client-auth-cert", "ech-key",
		"allow-insecure", "mux-option",
	)},
	"hysteria2": {Protocol: "hysteria2", Fields: listenerFields(
		"users", "up", "down", "ignore-client-bandwidth", "obfs", "obfs-password", "masquerade",
		"bbr-profile", "alpn", "certificate", "private-key", "client-auth-type", "client-auth-cert",
		"ech-key", "allow-insecure", "mux-option",
	)},
	"tuic": {Protocol: "tuic", Fields: listenerFields(
		"users", "token", "certificate", "private-key", "client-auth-type", "client-auth-cert", "ech-key",
		"allow-insecure", "congestion-controller", "bbr-profile", "max-idle-time", "authentication-timeout",
		"alpn", "max-udp-relay-packet-size", "mux-option",
	)},
	"shadowquic": {Protocol: "shadowquic", Fields: listenerFields(
		"users", "jls-upstream", "alpn", "quic-versions", "zero-rtt", "congestion-controller", "up", "down",
		"ignore-client-bandwidth", "cwnd", "bbr-profile", "max-idle-time", "max-datagram-frame-size",
		"recv-window-conn", "recv-window", "disable-mtu-discovery",
	)},
	"anytls": {Protocol: "anytls", Fields: listenerFields(
		"users", "certificate", "private-key", "client-auth-type", "client-auth-cert", "ech-key",
		"allow-insecure", "padding-scheme",
	)},
	"mieru": {Protocol: "mieru", Fields: listenerFields(
		"transport", "users", "traffic-pattern", "user-hint-is-mandatory",
	)},
	"sudoku": {Protocol: "sudoku", Fields: listenerFields(
		"key", "aead-method", "padding-min", "padding-max", "table-type", "custom-table", "custom-tables",
		"handshake-timeout", "enable-pure-downlink", "httpmask", "fallback", "mux-option",
	)},
	"trusttunnel": {Protocol: "trusttunnel", Fields: listenerFields(
		"users", "certificate", "private-key", "client-auth-type", "client-auth-cert", "ech-key", "network",
		"congestion-controller", "bbr-profile",
	)},
}

func GetMihomoListenerSchema(protocol string) (ListenerSchema, bool) {
	schema, ok := MihomoListenerSchemas[protocol]
	return schema, ok
}
