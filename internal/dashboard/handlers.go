package dashboard

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"scraper/internal/storage"
)

// handleHome serves the main dashboard page
func (d *Dashboard) handleHome(w http.ResponseWriter, r *http.Request) {
	tmplParsed, err := template.New("dashboard").Parse(MainTemplate)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	tmplParsed.Execute(w, nil)
}

// handleAPIContracts returns contracts as JSON
func (d *Dashboard) handleAPIContracts(w http.ResponseWriter, r *http.Request) {
	contracts, err := d.store.GetContracts()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get contracts: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(contracts)
}

// handleAPIStats returns statistics as JSON
func (d *Dashboard) handleAPIStats(w http.ResponseWriter, r *http.Request) {
	count, err := d.store.GetContractCount()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get stats: %v", err), http.StatusInternalServerError)
		return
	}

	stats := map[string]interface{}{
		"total":    count,
		"newToday": 0, // TODO: Implement new today logic
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// handleDeleteAll deletes all contracts
func (d *Dashboard) handleDeleteAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := d.store.DeleteAllContracts()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

// handleDeleteContract deletes a specific contract
func (d *Dashboard) handleDeleteContract(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the request body
	var request struct {
		ID string `json:"id"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.ID == "" {
		http.Error(w, "Contract ID is required", http.StatusBadRequest)
		return
	}

	err := d.store.DeleteContract(request.ID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

// handleAPIStatusChanges returns recent status changes as JSON
func (d *Dashboard) handleAPIStatusChanges(w http.ResponseWriter, r *http.Request) {
	statusChanges, err := d.store.GetRecentStatusChanges()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get status changes: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(statusChanges)
}

// handleHistory displays the complete status changes history
func (d *Dashboard) handleHistory(w http.ResponseWriter, r *http.Request) {
	statusChanges, err := d.store.GetAllStatusChanges()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	tmplParsed, err := template.New("history").Parse(HistoryTemplate)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	
	data := struct {
		StatusChanges []storage.StatusChange
	}{
		StatusChanges: statusChanges,
	}
	
	w.Header().Set("Content-Type", "text/html")
	tmplParsed.Execute(w, data)
} 