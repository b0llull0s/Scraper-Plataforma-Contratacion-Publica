package dashboard

import (
	"log"
	"net/http"

	"scraper/internal/storage"
)

// Dashboard handles the web interface
type Dashboard struct {
	store *storage.Storage
	port  string
}

// NewDashboard creates a new dashboard instance
func NewDashboard(store *storage.Storage, port string) *Dashboard {
	return &Dashboard{
		store: store,
		port:  port,
	}
}

// Start starts the web server
func (d *Dashboard) Start() error {
	// Register all routes
	d.registerRoutes()

	addr := ":" + d.port
	log.Printf("Dashboard starting on http://localhost%s", addr)
	return http.ListenAndServe(addr, nil)
} 