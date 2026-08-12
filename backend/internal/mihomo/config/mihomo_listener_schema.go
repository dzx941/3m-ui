package config

// MihomoListenerProtocols is the set of real listener protocols exposed by
// the 3m-ui listener UI. Local proxy endpoints (socks/http/tproxy/redir/mixed),
// tunnel and TUN are intentionally not part of the node/client distribution
// model.
var MihomoListenerProtocols = []string{
	"shadowsocks",
	"snell",
	"vmess",
	"vless",
	"trojan",
	"hysteria2",
	"hysteria2-realm",
	"tuic",
	"shadowquic",
	"anytls",
	"mieru",
	"sudoku",
	"trusttunnel",
}

func IsMihomoListenerProtocol(protocol string) bool {
	for _, p := range MihomoListenerProtocols {
		if p == protocol {
			return true
		}
	}
	return false
}
