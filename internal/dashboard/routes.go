package dashboard

import "net/http"

// registerRoutes registers all HTTP routes for the dashboard
func (d *Dashboard) registerRoutes() {
	// Main pages
	http.HandleFunc("/", d.handleHome)
	http.HandleFunc("/history", d.handleHistory)
	
	// API endpoints
	http.HandleFunc("/api/contracts", d.handleAPIContracts)
	http.HandleFunc("/api/stats", d.handleAPIStats)
	http.HandleFunc("/api/delete-all", d.handleDeleteAll)
	http.HandleFunc("/api/delete-contract", d.handleDeleteContract)
	http.HandleFunc("/api/status-changes", d.handleAPIStatusChanges)
} 