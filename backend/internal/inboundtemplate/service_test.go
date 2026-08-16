package inboundtemplate

import (
	"encoding/json"
	"testing"

	"github.com/kazeyukiro/3m-ui/backend/internal/database/models"
	"github.com/kazeyukiro/3m-ui/backend/internal/user"
)

func TestListIncludesClashMetaTemplates(t *testing.T) {
	items := List()
	if len(items) < 6 { t.Fatalf("expected at least 6 templates, got %d", len(items)) }
	if _, ok := Find("vless-reality-tcp-vision"); !ok { t.Fatal("Reality template missing") }
	if _, ok := Find("hysteria2-ech"); !ok { t.Fatal("Hysteria2 template missing") }
}

func TestCreateVLESSReality(t *testing.T) {
	var saved *models.Listener
	var created *models.ProxyUser
	listener, proxyUser, err := Create(CreateInput{TemplateID:"vless-reality-tcp-vision", Port:"443", Username:"alice", Values:map[string]string{"dest":"example.com:443","server_name":"example.com"}}, func(l *models.Listener) error { saved=l; l.ID=1; return nil }, func(in user.CreateInput) (*models.ProxyUser,error) { created=&models.ProxyUser{ID:2,Username:in.Username,UUID:in.UUID}; return created,nil }, func(uint,[]uint) error { return nil })
	if err != nil { t.Fatal(err) }
	if listener != saved || proxyUser != created { t.Fatal("unexpected returned objects") }
	var cfg map[string]interface{}
	if err := json.Unmarshal([]byte(saved.Config), &cfg); err != nil { t.Fatal(err) }
	if cfg["reality-config"] == nil { t.Fatal("missing reality config") }
	if cfg["users"] == nil { t.Fatal("missing users") }
}

func TestPasswordBasedTemplatesRequirePassword(t *testing.T) {
	_, _, err := Create(CreateInput{TemplateID:"shadowquic", Values:map[string]string{"jls_addr":"example.com:443","jls_sni":"example.com"}}, func(*models.Listener) error { t.Fatal("listener should not be created"); return nil }, func(user.CreateInput) (*models.ProxyUser,error) { t.Fatal("user should not be created"); return nil,nil }, func(uint,[]uint) error { return nil })
	if err == nil { t.Fatal("expected password validation error") }
}
