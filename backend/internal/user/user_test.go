package user

import (
	"testing"
	"time"

	"github.com/dzx941/3m-ui/backend/internal/database"
	"github.com/dzx941/3m-ui/backend/internal/database/models"
	"github.com/dzx941/3m-ui/backend/internal/security"
)

func TestActiveCredentialsFiltering(t *testing.T) {
	security.InitCredentialKey("test-secret")
	db, err := database.InitDB(t.TempDir() + "/test.db")
	if err != nil { t.Fatal(err) }
	InitService(db)

	l1 := models.Listener{Name: "ss", Protocol: "shadowsocks", Port: 8388, Enabled: true}
	if err := db.Create(&l1).Error; err != nil { t.Fatal(err) }

	active, err := GlobalService.Create(CreateInput{Username: "active", Password: "p1"})
	if err != nil { t.Fatal(err) }
	expiredAt := time.Now().Add(-time.Hour)
	expired, err := GlobalService.Create(CreateInput{Username: "expired", Password: "p2", ExpireTime: &expiredAt})
	if err != nil { t.Fatal(err) }
	limited, err := GlobalService.Create(CreateInput{Username: "limited", Password: "p3", TrafficLimit: 10})
	if err != nil { t.Fatal(err) }
	if err := db.Model(limited).Update("traffic_used", int64(10)).Error; err != nil { t.Fatal(err) }

	for _, id := range []uint{active.ID, expired.ID, limited.ID} {
		if err := db.Create(&models.ListenerUser{ListenerID: l1.ID, ProxyUserID: id}).Error; err != nil { t.Fatal(err) }
	}

	creds, err := GlobalService.ActiveCredentialsByListener()
	if err != nil { t.Fatal(err) }
	if len(creds[l1.ID]) != 1 || creds[l1.ID][0].Password != "p1" {
		t.Fatalf("expected only active user credential, got %#v", creds[l1.ID])
	}
}

func TestBindListenersIsReplacement(t *testing.T) {
	security.InitCredentialKey("test-secret")
	db, err := database.InitDB(t.TempDir() + "/test.db")
	if err != nil { t.Fatal(err) }
	InitService(db)

	u, err := GlobalService.Create(CreateInput{Username: "bind-user", Password: "p"})
	if err != nil { t.Fatal(err) }
	var listeners []models.Listener
	for i := 0; i < 3; i++ {
		l := models.Listener{Name: "n", Protocol: "trojan", Port: 10000 + i, Enabled: true}
		if err := db.Create(&l).Error; err != nil { t.Fatal(err) }
		listeners = append(listeners, l)
	}
	if err := GlobalService.BindListeners(u.ID, []uint{listeners[0].ID, listeners[1].ID}); err != nil { t.Fatal(err) }
	if err := GlobalService.BindListeners(u.ID, []uint{listeners[2].ID}); err != nil { t.Fatal(err) }
	got, err := GlobalService.GetListeners(u.ID)
	if err != nil { t.Fatal(err) }
	if len(got) != 1 || got[0].ID != listeners[2].ID { t.Fatalf("expected replacement binding, got %#v", got) }
}
