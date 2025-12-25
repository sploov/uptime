package models

import (
	"time"
)

type ServiceStatus string

const (
	StatusOperational ServiceStatus = "operational"
	StatusDegraded    ServiceStatus = "degraded"
	StatusOutage      ServiceStatus = "outage"
)

type HeartbeatValue int

const (
	HeartbeatUp       HeartbeatValue = 0
	HeartbeatDegraded HeartbeatValue = 1
	HeartbeatDown     HeartbeatValue = 2
)

// ServiceConfig represents a target to monitor from config
type ServiceConfig struct {
	ID       string        `yaml:"id" json:"id"`
	Name     string        `yaml:"name" json:"name"`
	URL      string        `yaml:"url" json:"url"`
	Interval time.Duration `yaml:"interval" json:"interval"`
	Timeout  time.Duration `yaml:"timeout" json:"timeout"`
	Method   string        `yaml:"method" json:"method"` // GET, TCP, ICMP (simulated via ping or just TCP connect)
}

// CheckResult represents the result of a single poll
type CheckResult struct {
	ServiceID   string         `json:"service_id"`
	Timestamp   time.Time      `json:"timestamp"`
	Latency     time.Duration  `json:"latency_ns"`
	Status      HeartbeatValue `json:"status"`
	Error       string         `json:"error,omitempty"`
}

// Service is the JSON API response format
type Service struct {
	ID         string        `json:"id"`
	Name       string        `json:"name"`
	Uptime     string        `json:"uptime"` // e.g. '99.98%'
	Status     ServiceStatus `json:"status"`
	Latency    float64       `json:"latency"` // in ms
	Heartbeats []int         `json:"heartbeats"` // Last N heartbeats
}
