# Sploov Uptime Engine

![Sploov Uptime](https://img.shields.io/badge/Status-Operational-green) ![Go Version](https://img.shields.io/github/go-mod/go-version/sploov/uptime)

A high-performance, distributed monitoring system written in Go. It performs synthetic protocol handshakes (HTTP, TCP) across your infrastructure and exposes that data via a clean JSON API and a built-in dashboard.

**Features:**
*   **Concurrent Polling**: Checks dozens of services simultaneously using lightweight Goroutines.
*   **Dynamic Configuration**: Hot-reloadable target list via `config.yaml`.
*   **History Persistence**: Built-in SQLite storage for up to 90 days of metrics.
*   **Alerting**: Native Discord Webhook integration.
*   **Modern Dashboard**: A clean, responsive dark-mode UI to view your service status in real-time.
*   **Docker Ready**: One-command deployment.

---

## 🚀 Quick Start

### 1. Using Docker (Recommended)

```bash
# Clone the repo
git clone https://github.com/sploov/uptime.git
cd uptime

# Start the engine
docker-compose up -d
```

Visit **http://localhost:8080** to see the dashboard.

### 2. Manual Installation

**Prerequisites:** Go 1.25+

```bash
# Get dependencies
go mod tidy

# Build the binary
go build -o uptime-engine cmd/server/main.go

# Run
./uptime-engine
```

## ⚙️ Configuration

Edit `config.yaml` to define your monitored services.

```yaml
targets:
  - id: "my-api"
    name: "Production API"
    url: "https://api.example.com/health"
    method: "HTTP"
    interval: 30s
    timeout: 5s

  - id: "db-primary"
    name: "Primary Database"
    url: "db.example.com:5432"
    method: "TCP"
    interval: 10s

discord:
  enabled: true
  webhook_url: "https://discord.com/api/webhooks/..."
```

## 📡 API Documentation

### `GET /api/status`
Returns the current health of all services.

```json
[
  {
    "id": "google",
    "name": "Google",
    "uptime": "99.99%",
    "status": "operational",
    "latency": 45,
    "heartbeats": [0, 0, 0, 1, 0] 
  }
]
```
*Note: Heartbeat values: 0=Up, 1=Degraded, 2=Down*

### `GET /api/history/{id}`
Returns raw check history for a specific service.

## 🛠️ Development

### Project Structure
*   `cmd/server`: Application entry point.
*   `internal/monitor`: Core concurrent monitoring logic.
*   `internal/storage`: SQLite persistence layer.
*   `internal/api`: REST API and Dashboard handlers.

### Building for Release
Use the Makefile (if available) or standard Go build commands.
```bash
CGO_ENABLED=1 go build -o uptime-engine cmd/server/main.go
```

## 📄 License
MIT License. See [LICENSE](LICENSE) for details.