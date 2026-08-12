package converter

import (
	"testing"

	"github.com/dzx941/3m-ui/backend/internal/database/models"
)

func TestListenerToProxyConvertsServerFields(t *testing.T) {
	listener := models.Listener{
		Name: "vless-in",
		Protocol: "vless",
		Port: 443,
		BindAddress: "0.0.0.0",
		TLS: true,
		Config: `{"flow":"xtls-rprx-vision","ws-path":"/ws","grpc-service-name":"grpc","certificate":"server.crt","private-key":"server.key","reality-config":{"private-key":"server-secret","public-key":"client-public","short-id":["0123456789abcdef"]}}`,
	}

	proxy, err := listenerToProxy(listener, "alice", "password", "11111111-1111-4111-8111-111111111111", "example.com")
	if err != nil { t.Fatalf("listenerToProxy failed: %v", err) }
	if proxy["server"] != "example.com" || proxy["port"] != 443 { t.Fatal("server endpoint was not converted") }
	if proxy["tls"] != true { t.Fatal("TLS was not enabled for client export") }
	if proxy["certificate"] != nil || proxy["private-key"] != nil || proxy["reality-config"] != nil { t.Fatal("server-only TLS secrets leaked into client proxy config") }
	if proxy["uuid"] != "11111111-1111-4111-8111-111111111111" { t.Fatal("uuid was not exported") }
	if proxy["flow"] != "xtls-rprx-vision" { t.Fatal("flow was not exported") }
	ws := proxy["ws-opts"].(map[string]interface{})
	if ws["path"] != "/ws" { t.Fatal("websocket path was not converted") }
	reality := proxy["reality-opts"].(map[string]interface{})
	if reality["public-key"] != "client-public" || reality["short-id"] != "0123456789abcdef" { t.Fatal("reality client options were not converted") }
}
