package config

// ListenerSchema describes only the protocol-specific fields accepted inside
// a Mihomo listener. Common listener fields (name, type, port, listen, rule,
// proxy) are generated separately and are never part of this registry.
type ListenerSchema struct {
	Protocol      string
	Fields        map[string]struct{}
	NestedFields  map[string]map[string]struct{}
}

func listenerFields(values ...string) map[string]struct{} {
	fields := make(map[string]struct{}, len(values))
	for _, value := range values {
		fields[value] = struct{}{}
	}
	return fields
}

func listenerNested(values ...string) map[string]map[string]struct{} {
	fields := make(map[string]map[string]struct{})
	for _, value := range values {
		parts := splitPath(value)
		if len(parts) < 2 {
			continue
		}
		fields[parts[0]] = appendNested(fields[parts[0]], parts[1:])
	}
	return fields
}

func appendNested(current map[string]struct{}, path []string) map[string]struct{} {
	if current == nil {
		current = make(map[string]struct{})
	}
	if len(path) > 0 {
		current[path[0]] = struct{}{}
	}
	return current
}

func splitPath(value string) []string {
	var out []string
	start := 0
	for i := 0; i <= len(value); i++ {
		if i == len(value) || value[i] == '.' {
			if i > start {
				out = append(out, value[start:i])
			}
			start = i + 1
		}
	}
	return out
}

// MihomoListenerSchemas follows the current MetaCubeX listener documentation.
// Local traffic listeners (SOCKS/HTTP/Mixed/REDIR/TPROXY/TUN/Tunnel) are
// intentionally excluded because this registry is used for distributable
// proxy nodes and client-profile generation.
var MihomoListenerSchemas = map[string]ListenerSchema{
	"shadowsocks": {Protocol: "shadowsocks", Fields: listenerFields("cipher", "password", "udp", "shadow-tls", "kcp-tun"), NestedFields: listenerNested(
		"shadow-tls.enable", "shadow-tls.version", "shadow-tls.password", "shadow-tls.users", "shadow-tls.handshake",
		"kcp-tun.enable", "kcp-tun.key", "kcp-tun.crypt", "kcp-tun.mode", "kcp-tun.conn", "kcp-tun.autoexpire", "kcp-tun.scavengettl", "kcp-tun.ratelimit", "kcp-tun.mtu", "kcp-tun.sndwnd", "kcp-tun.rcvwnd", "kcp-tun.datashard", "kcp-tun.parityshard", "kcp-tun.dscp", "kcp-tun.nocomp", "kcp-tun.acknodelay", "kcp-tun.nodelay", "kcp-tun.interval", "kcp-tun.resend", "kcp-tun.sockbuf", "kcp-tun.smuxver", "kcp-tun.smuxbuf", "kcp-tun.framesize", "kcp-tun.streambuf", "kcp-tun.keepalive",
	)},
	"snell": {Protocol: "snell", Fields: listenerFields("psk", "version", "udp", "obfs-opts"), NestedFields: listenerNested("obfs-opts.mode", "obfs-opts.host")},
	"vmess": {Protocol: "vmess", Fields: listenerFields("users", "ws-path", "grpc-service-name", "mekya-config", "mkcp-config", "shadow-tls", "res-tls", "jls-config", "reality-config", "tlsmirror-config", "certificate", "private-key", "client-auth-type", "client-auth-cert", "ech-key", "allow-insecure"), NestedFields: listenerNested(
		"mekya-config.enable", "mekya-config.max-write-size", "mekya-config.max-write-duration-ms", "mekya-config.max-simultaneous-write-connection", "mekya-config.packet-writing-buffer",
		"mkcp-config.mtu", "mkcp-config.tti", "mkcp-config.uplink-capacity", "mkcp-config.downlink-capacity", "mkcp-config.congestion", "mkcp-config.write-buffer", "mkcp-config.read-buffer", "mkcp-config.seed", "mkcp-config.header",
		"reality-config.dest", "reality-config.private-key", "reality-config.short-id", "reality-config.server-names", "reality-config.limit-fallback-upload", "reality-config.limit-fallback-download",
		"tlsmirror-config.dest", "tlsmirror-config.primary-key",
	)},
	"vless": {Protocol: "vless", Fields: listenerFields("users", "ws-path", "grpc-service-name", "xhttp-config", "decryption", "reality-config", "certificate", "private-key", "client-auth-type", "client-auth-cert", "ech-key", "allow-insecure"), NestedFields: listenerNested(
		"xhttp-config.path", "xhttp-config.host", "xhttp-config.mode", "xhttp-config.no-sse-header", "xhttp-config.x-padding-bytes", "xhttp-config.x-padding-obfs-mode", "xhttp-config.x-padding-key", "xhttp-config.x-padding-header", "xhttp-config.x-padding-placement", "xhttp-config.x-padding-method", "xhttp-config.uplink-http-method", "xhttp-config.session-placement", "xhttp-config.session-key", "xhttp-config.session-table", "xhttp-config.session-length", "xhttp-config.seq-placement", "xhttp-config.seq-key", "xhttp-config.uplink-data-placement", "xhttp-config.uplink-data-key", "xhttp-config.uplink-chunk-size", "xhttp-config.sc-max-buffered-posts", "xhttp-config.sc-stream-up-server-secs", "xhttp-config.sc-max-each-post-bytes",
		"reality-config.dest", "reality-config.private-key", "reality-config.short-id", "reality-config.server-names", "reality-config.limit-fallback-upload", "reality-config.limit-fallback-download",
	)},
	"trojan": {Protocol: "trojan", Fields: listenerFields("users", "ws-path", "grpc-service-name", "reality-config", "ss-option", "certificate", "private-key", "client-auth-type", "client-auth-cert", "ech-key", "allow-insecure"), NestedFields: listenerNested(
		"reality-config.dest", "reality-config.private-key", "reality-config.short-id", "reality-config.server-names", "reality-config.limit-fallback-upload", "reality-config.limit-fallback-download",
		"ss-option.enabled", "ss-option.method", "ss-option.password",
	)},
	"hysteria2": {Protocol: "hysteria2", Fields: listenerFields("users", "up", "down", "ignore-client-bandwidth", "obfs", "obfs-password", "masquerade", "realm-opts", "bbr-profile", "alpn", "certificate", "private-key", "client-auth-type", "client-auth-cert", "ech-key"), NestedFields: listenerNested(
		"realm-opts.enable", "realm-opts.server-url", "realm-opts.token", "realm-opts.realm-id", "realm-opts.stun-servers", "realm-opts.proxy", "realm-opts.skip-cert-verify", "realm-opts.fingerprint", "realm-opts.certificate", "realm-opts.private-key", "realm-opts.alpn",
	)},
	"tuic": {Protocol: "tuic", Fields: listenerFields("users", "token", "certificate", "private-key", "client-auth-type", "client-auth-cert", "ech-key", "congestion-controller", "bbr-profile", "max-idle-time", "authentication-timeout", "alpn", "max-udp-relay-packet-size")},
	"shadowquic": {Protocol: "shadowquic", Fields: listenerFields("users", "jls-upstream", "alpn", "quic-versions", "zero-rtt", "congestion-controller", "up", "down", "ignore-client-bandwidth", "cwnd", "bbr-profile", "max-idle-time", "max-datagram-frame-size", "recv-window-conn", "recv-window", "disable-mtu-discovery")},
	"anytls": {Protocol: "anytls", Fields: listenerFields("users", "certificate", "private-key", "client-auth-type", "client-auth-cert", "ech-key", "allow-insecure", "padding-scheme")},
	"mieru": {Protocol: "mieru", Fields: listenerFields("transport", "users", "traffic-pattern", "user-hint-is-mandatory")},
	"sudoku": {Protocol: "sudoku", Fields: listenerFields("key", "aead-method", "padding-min", "padding-max", "table-type", "custom-table", "custom-tables", "handshake-timeout", "enable-pure-downlink", "httpmask", "fallback", "disable-http-mask"), NestedFields: listenerNested(
		"httpmask.disable", "httpmask.mode", "httpmask.path_root", "httpmask.tls", "httpmask.host", "httpmask.path-root", "httpmask.multiplex",
	)},
	"trusttunnel": {Protocol: "trusttunnel", Fields: listenerFields("users", "certificate", "private-key", "client-auth-type", "client-auth-cert", "ech-key", "network", "congestion-controller", "bbr-profile")},
}

func GetMihomoListenerSchema(protocol string) (ListenerSchema, bool) {
	schema, ok := MihomoListenerSchemas[protocol]
	return schema, ok
}
