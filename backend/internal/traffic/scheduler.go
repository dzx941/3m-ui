package traffic

import (
	"log"
	"sync"
	"time"
)

// DefaultInterval is how often the scheduler polls Mihomo when no interval
// is explicitly configured.
const DefaultInterval = 10 * time.Second

// Scheduler runs the traffic collector and enforcer on a fixed interval.
// It is started once at server startup (see cmd/server/main.go) and
// stopped gracefully on shutdown.
type Scheduler struct {
	collector *Collector
	enforcer  *Enforcer
	interval  time.Duration

	stopCh chan struct{}
	wg     sync.WaitGroup
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
//  2. update global statistics + TrafficRecord history (done inside CollectOnce)
//  3. update per-user counters/online state (done inside CollectOnce)
//  4. check limits and enforce (Enforcer.CheckAndEnforce)
func (s *Scheduler) Start() {
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
		// condition (e.g. before the admin has started it), so this is a
		// log rather than a panic/fatal.
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

// Stop signals the collection loop to exit and blocks until it has, for
// graceful shutdown.
func (s *Scheduler) Stop() {
	close(s.stopCh)
	s.wg.Wait()
}
