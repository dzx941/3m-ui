package traffic

import (
	"context"
	"time"
)

type Scheduler struct {
	collector *Collector
	interval  time.Duration
	stopChan  chan struct{}
	running   bool
}

var GlobalScheduler *Scheduler

func InitScheduler(configPath string, service *Service) {
	collector := NewCollector(configPath, service)
	GlobalScheduler = &Scheduler{
		collector: collector,
		interval:  10 * time.Second,
		stopChan:  make(chan struct{}),
	}
}

func (s *Scheduler) Start() {
	if s.running {
		return
	}
	s.running = true
	go func() {
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		// Initial collection immediately
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		_ = s.collector.Collect(ctx)
		cancel()

		for {
			select {
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
				_ = s.collector.Collect(ctx)
				cancel()
			case <-s.stopChan:
				return
			}
		}
	}()
}

func (s *Scheduler) Stop() {
	if !s.running {
		return
	}
	close(s.stopChan)
	s.running = false
}
