package config

// ListenerSchema describes only the protocol-specific fields accepted inside
// a Mihomo listener. Common listener fields (name, type, port, listen, rule,
// proxy) are generated separately and are never part of this registry.
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

// MihomoListenerSchemas follows the current MetaCubeX listener documentation.
// Local traffic listeners (SOCKS/HTTP/Mixed/REDIR/TPROXY/TUN/Tunnel) are
// intentionally excluded because this registry is used for distributable
// proxy nodes and client-profile generation.
var MihomoListenerSchemas = map[string]ListenerSchema{
	"shadowsocks": {Protocol: "shadowsocks", Fields: listenerFields(
		"cipher", "password", "udp", "shadow-tls", "kcp-tun",
	)},
	"snell": {Protocol: "snell", Fields: listenerFields(
		"psk", "version", "udp", "obfs-opts",
	)},
	"vmess": {Protocol: "vmess", Fields: listenerFields(
		"users", "ws-path", "grpc-service-name", "mekya-config", "mkcp-config",
		"shadow-tls", "res-tls", "jls-config", "reality-config", "tlsmirror-config",
		"certificate", "private-key", "client-auth-type", "client-auth-cert", "ech-key",
		"allow-insecure",
	)},
	"vless": {Protocol: "vless", Fields: listenerFields(
		"users", "ws-path", "grpc-service-name", "xhttp-config", "decryption",
		"reality-config", "certificate", "private-key", "client-auth-type",
		"client-auth-cert", "ech-key", "allow-insecure",
	)},
	"trojan": {Protocol: "trojan", Fields: listenerFields(
		"users", "ws-path", "grpc-service-name", "reality-config", "ss-option",
		"certificate", "private-key", "client-auth-type", "client-auth-cert",
		"ech-key", "allow-insecure",
	)},
	"hysteria2": {Protocol: "hysteria2", Fields: listenerFields(
		"users", "up", "down", "ignore-client-bandwidth", "obfs", "obfs-password",
		"masquerade", "realm-opts", "bbr-profile", "alpn", "certificate", "private-key",
		"client-auth-type", "client-auth-cert", "ech-key",
	)},
	"tuic": {Protocol: "tuic", Fields: listenerFields(
		"users", "token", "certificate", "private-key", "client-auth-type",
		"client-auth-cert", "ech-key", "congestion-controller", "bbr-profile",
		"max-idle-time", "authentication-timeout", "alpn", "max-udp-relay-packet-size",
	)},
	"shadowquic": {Protocol: "shadowquic", Fields: listenerFields(
		"users", "jls-upstream", "alpn", "quic-versions", "zero-rtt",
		"congestion-controller", "up", "down", "ignore-client-bandwidth", "cwnd",
		"bbr-profile", "max-idle-time", "max-datagram-frame-size", "recv-window-conn",
		"recv-window", "disable-mtu-discovery",
	)},
	"anytls": {Protocol: "anytls", Fields: listenerFields(
		"users", "certificate", "private-key", "client-auth-type", "client-auth-cert",
		"ech-key", "allow-insecure", "padding-scheme",
	)},
	"mieru": {Protocol: "mieru", Fields: listenerFields(
		"transport", "users", "traffic-pattern", "user-hint-is-mandatory",
	)},
	"sudoku": {Protocol: "sudoku", Fields: listenerFields(
		"key", "aead-method", "padding-min", "padding-max", "table-type", "custom-table",
		"custom-tables", "handshake-timeout", "enable-pure-downlink", "httpmask",
		"fallback", "disable-http-mask",
	)},
	"trusttunnel": {Protocol: "trusttunnel", Fields: listenerFields(
		"users", "certificate", "private-key", "client-auth-type", "client-auth-cert",
		"ech-key", "network", "congestion-controller", "bbr-profile",
	)},
}

func GetMihomoListenerSchema(protocol string) (ListenerSchema, bool) {
	schema, ok := MihomoListenerSchemas[protocol]
	return schema, ok
}
