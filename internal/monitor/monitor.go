package monitor

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/sploov/uptime/internal/alert"
	"github.com/sploov/uptime/internal/config"
	"github.com/sploov/uptime/internal/models"
	"github.com/sploov/uptime/internal/storage"
)

type Monitor struct {
	cfg      *config.Config
	store    *storage.Store
	notifier alert.Notifier

	// In-memory cache for current status to serve API fast
	currentMu sync.RWMutex
	current   map[string]*models.Service // Keyed by ID
}

func NewMonitor(cfg *config.Config, store *storage.Store, notifier alert.Notifier) *Monitor {
	m := &Monitor{
		cfg:      cfg,
		store:    store,
		notifier: notifier,
		current:  make(map[string]*models.Service),
	}
	// Initialize cache
	for _, t := range cfg.Targets {
		m.current[t.ID] = &models.Service{
			ID:     t.ID,
			Name:   t.Name,
			Status: models.StatusOperational, // Assume up initially or unknown
		}
	}
	return m
}

func (m *Monitor) Start(ctx context.Context) {
	for _, target := range m.cfg.Targets {
		go m.pollLoop(ctx, target)
	}
}

func (m *Monitor) pollLoop(ctx context.Context, target models.ServiceConfig) {
	ticker := time.NewTicker(target.Interval)
	defer ticker.Stop()

	// Initial check
	m.check(target)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.check(target)
		}
	}
}

func (m *Monitor) check(target models.ServiceConfig) {
	start := time.Now()
	var err error
	var status models.HeartbeatValue = models.HeartbeatUp

	// Perform the check
	switch target.Method {
	case "HTTP", "HTTPS":
		err = checkHTTP(target.URL, target.Timeout)
	case "TCP":
		err = checkTCP(target.URL, target.Timeout)
	default:
		// Fallback to TCP if unknown, or handle ICMP if implemented
		err = checkTCP(target.URL, target.Timeout)
	}

	latency := time.Since(start)
	if err != nil {
		status = models.HeartbeatDown
		// Simple retry logic could go here, but for now we just mark down
	} else if latency > target.Timeout/2 {
		// Example logic for degraded: if latency is high (e.g. > 50% of timeout)
		status = models.HeartbeatDegraded
	}

	// Determine service status
	srvStatus := models.StatusOperational
	if status == models.HeartbeatDown {
		srvStatus = models.StatusOutage
	} else if status == models.HeartbeatDegraded {
		srvStatus = models.StatusDegraded
	}

	// Update storage
	checkResult := models.CheckResult{
		ServiceID: target.ID,
		Timestamp: start,
		Latency:   latency,
		Status:    status,
	}
	if err != nil {
		checkResult.Error = err.Error()
	}
	m.store.AddCheck(checkResult)

	// Update in-memory cache and notify if changed
	m.currentMu.Lock()
	cached, exists := m.current[target.ID]
	if !exists {
		// Should not happen if initialized, but handle safety
		cached = &models.Service{ID: target.ID, Name: target.Name}
		m.current[target.ID] = cached
	}

	oldStatus := cached.Status
	cached.Status = srvStatus
	cached.Latency = float64(latency.Milliseconds())
	// We don't store heartbeats array in memory permanently, we fetch from DB for API
	// But we could update a ring buffer here if we wanted to avoid DB hits for "recent"
	m.currentMu.Unlock()

	if oldStatus != srvStatus {
		// Status changed, notify
		// We might want to debounce this (only notify if down for 2 checks), but keeping it simple
		go m.notifier.Notify(target.Name, target.URL, oldStatus, srvStatus, latency)
	}
}

func checkHTTP(url string, timeout time.Duration) error {
	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP status %d", resp.StatusCode)
	}
	return nil
}

func checkTCP(address string, timeout time.Duration) error {
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}

// GetServices returns the current status of all services
func (m *Monitor) GetServices() []models.Service {
	m.currentMu.RLock()
	defer m.currentMu.RUnlock()

	var services []models.Service
	for _, s := range m.current {
		// Create a copy
		svc := *s
		
		// Populate dynamic data (Uptime, Heartbeats) from DB
		// Note: Doing this inside the lock/loop might be slow if we have many services. 
		// For high performance, we should compute these async or cache them.
		// For now, we will fetch them here.
		
		// 1. Calculate Uptime (last 24h or 90d?) - Req says 90 days storage. 
		// Let's show uptime for last 24h in the list for relevance? 
		// Or 90d. Let's do 24h for "Current Status" view usually.
		uptime, _ := m.store.GetUptime(svc.ID, time.Now().Add(-24*time.Hour))
		svc.Uptime = fmt.Sprintf("%.2f%%", uptime)

		// 2. Get recent heartbeats (e.g. last 20 for a sparkline)
		checks, _ := m.store.GetRecentChecks(svc.ID, 20)
		var beats []int
		for _, c := range checks {
			beats = append(beats, int(c.Status))
		}
		svc.Heartbeats = beats
		
		services = append(services, svc)
	}
	return services
}

func (m *Monitor) GetServiceHistory(id string) ([]models.CheckResult, error) {
	// 100 recent checks
	return m.store.GetRecentChecks(id, 100)
}
