package traffic

import (
	"sync"
	"time"

	"github.com/dzx941/3m-ui/backend/internal/mihomo/api"
)

type Service struct {
	mu       sync.Mutex
	last     Snapshot
	lastTime time.Time
	conns    []api.Connection
}

var GlobalService *Service

func InitService() {
	GlobalService = NewService()
}

func NewService() *Service {
	return &Service{
		lastTime: time.Now(),
		conns:    make([]api.Connection, 0),
	}
}

func (s *Service) Update(totalUpload, totalDownload int64, connections int, apiConns []api.Connection) Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	result := Snapshot{
		UploadBytes:   totalUpload,
		DownloadBytes: totalDownload,
		Connections:   connections,
		UploadRate:    s.last.UploadRate,
		DownloadRate:  s.last.DownloadRate,
	}

	s.last = result
	s.lastTime = now
	s.conns = apiConns
	return result
}

func (s *Service) SetRates(upRate, downRate int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.last.UploadRate = upRate
	s.last.DownloadRate = downRate
}

func (s *Service) GetSnapshot() Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.last
}

func (s *Service) GetConnections() []api.Connection {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.conns == nil {
		return []api.Connection{}
	}
	copied := make([]api.Connection, len(s.conns))
	copy(copied, s.conns)
	return copied
}
