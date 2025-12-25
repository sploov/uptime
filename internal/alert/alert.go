package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sploov/uptime/internal/models"
)

type Notifier interface {
	Notify(serviceName, url string, oldStatus, newStatus models.ServiceStatus, latency time.Duration) error
}

type DiscordNotifier struct {
	WebhookURL string
}

func NewDiscordNotifier(url string) *DiscordNotifier {
	return &DiscordNotifier{WebhookURL: url}
}

func (d *DiscordNotifier) Notify(serviceName, url string, oldStatus, newStatus models.ServiceStatus, latency time.Duration) error {
	if d.WebhookURL == "" {
		return nil
	}

	color := 0x00FF00 // Green
	if newStatus == models.StatusDegraded {
		color = 0xFFFF00 // Yellow
	} else if newStatus == models.StatusOutage {
		color = 0xFF0000 // Red
	}

	payload := map[string]interface{}{
		"username": "Sploov Uptime",
		"embeds": []map[string]interface{}{
			{
				"title":       fmt.Sprintf("Status Change: %s", serviceName),
				"description": fmt.Sprintf("Service **%s** (%s) is now **%s**.\nPrevious status: %s\nLatency: %v", serviceName, url, newStatus, oldStatus, latency),
				"color":       color,
				"timestamp":   time.Now().Format(time.RFC3339),
			},
		},
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(d.WebhookURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("discord API returned status: %d", resp.StatusCode)
	}

	return nil
}
