package traffic

import (
	"log"
	"sync"
	"time"
)

// DefaultInterval is how often the scheduler polls Mihomo when no interval
// is explicitly configured.
const DefaultInterval = 10 * time.Second

// Scheduler owns the traffic collection loop and enforcement lifecycle.
// It is intentionally instance-scoped so the application container can
// control exactly one runtime scheduler instead of relying on package globals.
type Scheduler struct {
	collector *Collector
	enforcer  *Enforcer
	interval  time.Duration

	stopCh    chan struct{}
	stopOnce  sync.Once
	wg        sync.WaitGroup
}

// NewScheduler builds a Scheduler. A zero/negative interval falls back to
// DefaultInterval (10s).
func NewScheduler(collector *Collector, enforcer *Enforcer, interval time.Duration) *Scheduler {
	if interval <= 0 {
		interval = DefaultInterval
	}
	return &Scheduler{
		collector: collector,
		enforcer:  enforcer,
		interval:  interval,
		stopCh:    make(chan struct{}),
	}
}

// Start begins the collection loop in a background goroutine. Every tick:
//  1. collect Mihomo traffic (Collector.CollectOnce)
//  2. update traffic statistics and per-user state
//  3. check limits and enforce configuration changes
func (s *Scheduler) Start() {
	if s == nil || s.collector == nil {
		return
	}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		// Run an immediate first tick so dashboards aren't empty for a
		// full interval after startup.
		s.tick()

		for {
			select {
			case <-ticker.C:
				s.tick()
			case <-s.stopCh:
				return
			}
		}
	}()
}

func (s *Scheduler) tick() {
	if err := s.collector.CollectOnce(); err != nil {
		// Mihomo core being stopped/unreachable is an expected, recoverable
		// condition, so log and retry on the next tick.
		log.Printf("traffic: collection skipped: %v", err)
		return
	}
	if s.enforcer == nil {
		return
	}
	if _, err := s.enforcer.CheckAndEnforce(); err != nil {
		log.Printf("traffic: enforcement check failed: %v", err)
	}
}

// Stop signals the collection loop to exit and blocks until it has exited.
func (s *Scheduler) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() { close(s.stopCh) })
	s.wg.Wait()
}
