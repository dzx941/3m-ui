package traffic

import (
	"sync"
	"time"
)

type Service struct {
	mu sync.Mutex
	last Snapshot
	lastTime time.Time
}

func NewService() *Service {
	return &Service{lastTime: time.Now()}
}

func (s *Service) Update(totalUpload, totalDownload int64, connections int) Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	seconds := now.Sub(s.lastTime).Seconds()
	result := Snapshot{
		UploadBytes: totalUpload,
		DownloadBytes: totalDownload,
		Connections: connections,
	}

	if seconds > 0 {
		result.UploadRate = int64(float64(totalUpload-s.last.UploadBytes) / seconds)
		result.DownloadRate = int64(float64(totalDownload-s.last.DownloadBytes) / seconds)
	}

	s.last = result
	s.lastTime = now
	return result
}
