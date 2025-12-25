package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sploov/uptime/internal/alert"
	"github.com/sploov/uptime/internal/api"
	"github.com/sploov/uptime/internal/config"
	"github.com/sploov/uptime/internal/monitor"
	"github.com/sploov/uptime/internal/storage"
)

func main() {
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	port := flag.String("port", ":8080", "API server port")
	dbPath := flag.String("db", "uptime.db", "Path to SQLite database")
	flag.Parse()

	// 1. Load Configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Initialize Storage
	store, err := storage.NewStore(*dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()

	// 3. Initialize Notifier
	var notifier alert.Notifier
	if cfg.Discord.Enabled {
		notifier = alert.NewDiscordNotifier(cfg.Discord.WebhookURL)
	} else {
		notifier = &alert.DiscordNotifier{} // No-op if URL empty
	}

	// 4. Initialize Monitor
	mon := monitor.NewMonitor(cfg, store, notifier)

	// 5. Start Monitoring
	ctx, cancel := context.WithCancel(context.Background())
	mon.Start(ctx)

	// 6. Setup API
	handler := api.NewHandler(mon)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	srv := &http.Server{
		Addr:    *port,
		Handler: mux,
	}

	// 7. Run Server with Graceful Shutdown
	go func() {
		log.Printf("Starting Sploov Uptime Engine on %s", *port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down...")

	cancel() // Stop monitor

	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()
	if err := srv.Shutdown(ctxShutdown); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
