package storage

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"scraper/internal/scraper"
)

// Storage handles database operations
type Storage struct {
	db *sql.DB
}

// NewStorage creates a new storage instance
func NewStorage(dbPath string) (*Storage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	storage := &Storage{db: db}
	if err := storage.initTables(); err != nil {
		return nil, fmt.Errorf("failed to initialize tables: %w", err)
	}

	return storage, nil
}

// Close closes the database connection
func (s *Storage) Close() error {
	return s.db.Close()
}

// initTables creates the necessary tables if they don't exist
func (s *Storage) initTables() error {
	// Create contracts table
	contractsQuery := `
	CREATE TABLE IF NOT EXISTS contracts (
		id TEXT PRIMARY KEY,
		description TEXT,
		contract_type TEXT,
		status TEXT,
		amount TEXT,
		submission_date TEXT,
		contracting_body TEXT,
		link TEXT,
		pliego_link TEXT,
		anuncio_link TEXT,
		scraped_at DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err := s.db.Exec(contractsQuery)
	if err != nil {
		return fmt.Errorf("failed to create contracts table: %w", err)
	}

	// Create status changes table to track status modifications
	statusChangesQuery := `
	CREATE TABLE IF NOT EXISTS status_changes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		contract_id TEXT NOT NULL,
		old_status TEXT,
		new_status TEXT NOT NULL,
		changed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (contract_id) REFERENCES contracts (id)
	);
	`

	_, err = s.db.Exec(statusChangesQuery)
	if err != nil {
		return fmt.Errorf("failed to create status_changes table: %w", err)
	}

	log.Println("Database tables initialized successfully")
	return nil
}

// SaveContracts saves contracts to the database and tracks status changes
func (s *Storage) SaveContracts(contracts []scraper.Contract) error {
	if len(contracts) == 0 {
		return nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Prepare statements
	insertQuery := `
	INSERT OR REPLACE INTO contracts 
	(id, description, contract_type, status, amount, submission_date, contracting_body, link, pliego_link, anuncio_link, scraped_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`

	insertStmt, err := tx.Prepare(insertQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	defer insertStmt.Close()

	// Statement to check current status
	checkStatusQuery := `SELECT status FROM contracts WHERE id = ?`
	checkStatusStmt, err := tx.Prepare(checkStatusQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare check status statement: %w", err)
	}
	defer checkStatusStmt.Close()

	// Statement to insert status change
	statusChangeQuery := `INSERT INTO status_changes (contract_id, old_status, new_status) VALUES (?, ?, ?)`
	statusChangeStmt, err := tx.Prepare(statusChangeQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare status change statement: %w", err)
	}
	defer statusChangeStmt.Close()

	var statusChanges []string

	for _, contract := range contracts {
		// Check if contract exists and get current status
		var currentStatus string
		err := checkStatusStmt.QueryRow(contract.ID).Scan(&currentStatus)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("failed to check current status for contract %s: %w", contract.ID, err)
		}

		// Insert or update the contract
		_, err = insertStmt.Exec(
			contract.ID,
			contract.Description,
			contract.ContractType,
			contract.Status,
			contract.Amount,
			contract.SubmissionDate,
			contract.ContractingBody,
			contract.Link,
			contract.PliegoLink,
			contract.AnuncioLink,
			contract.ScrapedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert contract %s: %w", contract.ID, err)
		}

		// If contract existed and status changed, record the change
		if err != sql.ErrNoRows && currentStatus != "" && currentStatus != contract.Status {
			_, err = statusChangeStmt.Exec(contract.ID, currentStatus, contract.Status)
			if err != nil {
				return fmt.Errorf("failed to record status change for contract %s: %w", contract.ID, err)
			}
			statusChanges = append(statusChanges, fmt.Sprintf("%s: %s → %s", contract.ID, currentStatus, contract.Status))
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Saved %d contracts to database", len(contracts))
	if len(statusChanges) > 0 {
		log.Printf("Status changes detected: %v", statusChanges)
	}

	return nil
}

// CheckAndUpdateStatusChanges checks for status changes in existing contracts
// This method is called with ALL contracts found on the website to detect status changes
// for contracts that are already in our database but have different statuses
func (s *Storage) CheckAndUpdateStatusChanges(allContracts []scraper.Contract) error {
	if len(allContracts) == 0 {
		return nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Statement to check if contract exists and get current status
	checkQuery := `SELECT status FROM contracts WHERE id = ?`
	checkStmt, err := tx.Prepare(checkQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare check statement: %w", err)
	}
	defer checkStmt.Close()

	// Statement to update contract status
	updateQuery := `UPDATE contracts SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	updateStmt, err := tx.Prepare(updateQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare update statement: %w", err)
	}
	defer updateStmt.Close()

	// Statement to insert status change
	statusChangeQuery := `INSERT INTO status_changes (contract_id, old_status, new_status) VALUES (?, ?, ?)`
	statusChangeStmt, err := tx.Prepare(statusChangeQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare status change statement: %w", err)
	}
	defer statusChangeStmt.Close()

	var statusChanges []string

	for _, contract := range allContracts {
		// Check if contract exists in our database
		var currentStatus string
		err := checkStmt.QueryRow(contract.ID).Scan(&currentStatus)
		if err == sql.ErrNoRows {
			// Contract not in our database, skip (we only track existing contracts)
			continue
		} else if err != nil {
			return fmt.Errorf("failed to check contract %s: %w", contract.ID, err)
		}

		// If status changed, update it and record the change
		if currentStatus != contract.Status {
			_, err = updateStmt.Exec(contract.Status, contract.ID)
			if err != nil {
				return fmt.Errorf("failed to update status for contract %s: %w", contract.ID, err)
			}

			_, err = statusChangeStmt.Exec(contract.ID, currentStatus, contract.Status)
			if err != nil {
				return fmt.Errorf("failed to record status change for contract %s: %w", contract.ID, err)
			}

			statusChanges = append(statusChanges, fmt.Sprintf("%s: %s → %s", contract.ID, currentStatus, contract.Status))
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	if len(statusChanges) > 0 {
		log.Printf("Status changes detected: %v", statusChanges)
	}

	return nil
}

// GetContracts retrieves all contracts from the database
func (s *Storage) GetContracts() ([]scraper.Contract, error) {
	query := `SELECT id, description, contract_type, status, amount, submission_date, contracting_body, link, pliego_link, anuncio_link, scraped_at FROM contracts ORDER BY scraped_at DESC`
	
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query contracts: %w", err)
	}
	defer rows.Close()

	var contracts []scraper.Contract
	for rows.Next() {
		var contract scraper.Contract
		err := rows.Scan(
			&contract.ID,
			&contract.Description,
			&contract.ContractType,
			&contract.Status,
			&contract.Amount,
			&contract.SubmissionDate,
			&contract.ContractingBody,
			&contract.Link,
			&contract.PliegoLink,
			&contract.AnuncioLink,
			&contract.ScrapedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan contract: %w", err)
		}
		contracts = append(contracts, contract)
	}

	return contracts, nil
}

// GetContractByID retrieves a specific contract by ID
func (s *Storage) GetContractByID(id string) (*scraper.Contract, error) {
	query := `SELECT id, description, contract_type, status, amount, submission_date, contracting_body, link, pliego_link, anuncio_link, scraped_at FROM contracts WHERE id = ?`
	
	var contract scraper.Contract
	err := s.db.QueryRow(query, id).Scan(
		&contract.ID,
		&contract.Description,
		&contract.ContractType,
		&contract.Status,
		&contract.Amount,
		&contract.SubmissionDate,
		&contract.ContractingBody,
		&contract.Link,
		&contract.PliegoLink,
		&contract.AnuncioLink,
		&contract.ScrapedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get contract: %w", err)
	}

	return &contract, nil
}

// GetNewContracts returns contracts that don't exist in the database
func (s *Storage) GetNewContracts(contracts []scraper.Contract) ([]scraper.Contract, error) {
	var newContracts []scraper.Contract

	for _, contract := range contracts {
		exists, err := s.contractExists(contract.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to check if contract exists: %w", err)
		}
		if !exists {
			newContracts = append(newContracts, contract)
		}
	}

	return newContracts, nil
}

// contractExists checks if a contract with the given ID exists
func (s *Storage) contractExists(id string) (bool, error) {
	query := `SELECT COUNT(*) FROM contracts WHERE id = ?`
	
	var count int
	err := s.db.QueryRow(query, id).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check contract existence: %w", err)
	}

	return count > 0, nil
}

// DeleteAllContracts removes all contracts from the database
func (s *Storage) DeleteAllContracts() error {
	query := `DELETE FROM contracts`
	
	_, err := s.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to delete all contracts: %w", err)
	}

	log.Println("All contracts deleted from database")
	return nil
}

// DeleteContract removes a specific contract from the database
func (s *Storage) DeleteContract(contractID string) error {
	query := `DELETE FROM contracts WHERE id = ?`
	
	result, err := s.db.Exec(query, contractID)
	if err != nil {
		return fmt.Errorf("failed to delete contract %s: %w", contractID, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("contract %s not found", contractID)
	}

	log.Printf("Contract %s deleted from database", contractID)
	return nil
}

// GetContractCount returns the total number of contracts
func (s *Storage) GetContractCount() (int, error) {
	query := `SELECT COUNT(*) FROM contracts`
	
	var count int
	err := s.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get contract count: %w", err)
	}

	return count, nil
}

// StatusChange represents a status change record
type StatusChange struct {
	ID         int    `json:"id"`
	ContractID string `json:"contract_id"`
	OldStatus  string `json:"old_status"`
	NewStatus  string `json:"new_status"`
	ChangedAt  string `json:"changed_at"`
}

// GetStatusChanges retrieves all status changes for a specific contract
func (s *Storage) GetStatusChanges(contractID string) ([]StatusChange, error) {
	query := `
	SELECT id, contract_id, old_status, new_status, changed_at 
	FROM status_changes 
	WHERE contract_id = ? 
	ORDER BY changed_at DESC
	`
	
	rows, err := s.db.Query(query, contractID)
	if err != nil {
		return nil, fmt.Errorf("failed to query status changes: %w", err)
	}
	defer rows.Close()

	var changes []StatusChange
	for rows.Next() {
		var change StatusChange
		err := rows.Scan(
			&change.ID,
			&change.ContractID,
			&change.OldStatus,
			&change.NewStatus,
			&change.ChangedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan status change: %w", err)
		}
		changes = append(changes, change)
	}

	return changes, nil
}

// GetRecentStatusChanges retrieves recent status changes (last 24 hours)
func (s *Storage) GetRecentStatusChanges() ([]StatusChange, error) {
	query := `
	SELECT id, contract_id, old_status, new_status, changed_at 
	FROM status_changes 
	WHERE changed_at >= datetime('now', '-1 day')
	ORDER BY changed_at DESC
	`
	
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent status changes: %w", err)
	}
	defer rows.Close()

	var changes []StatusChange
	for rows.Next() {
		var change StatusChange
		err := rows.Scan(
			&change.ID,
			&change.ContractID,
			&change.OldStatus,
			&change.NewStatus,
			&change.ChangedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan status change: %w", err)
		}
		changes = append(changes, change)
	}

	return changes, nil
}

// GetAllStatusChanges retrieves all status changes
func (s *Storage) GetAllStatusChanges() ([]StatusChange, error) {
	query := `
	SELECT id, contract_id, old_status, new_status, changed_at 
	FROM status_changes 
	ORDER BY changed_at DESC
	`
	
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all status changes: %w", err)
	}
	defer rows.Close()

	var changes []StatusChange
	for rows.Next() {
		var change StatusChange
		err := rows.Scan(
			&change.ID,
			&change.ContractID,
			&change.OldStatus,
			&change.NewStatus,
			&change.ChangedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan status change: %w", err)
		}
		changes = append(changes, change)
	}

	return changes, nil
}

// GetContractsWithStatusChanges returns contracts that have recent status changes
func (s *Storage) GetContractsWithStatusChanges() ([]scraper.Contract, error) {
	query := `
	SELECT DISTINCT c.id, c.description, c.contract_type, c.status, c.amount, 
	       c.submission_date, c.contracting_body, c.scraped_at
	FROM contracts c
	INNER JOIN status_changes sc ON c.id = sc.contract_id
	WHERE sc.changed_at >= datetime('now', '-1 day')
	ORDER BY c.scraped_at DESC
	`
	
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query contracts with status changes: %w", err)
	}
	defer rows.Close()

	var contracts []scraper.Contract
	for rows.Next() {
		var contract scraper.Contract
		err := rows.Scan(
			&contract.ID,
			&contract.Description,
			&contract.ContractType,
			&contract.Status,
			&contract.Amount,
			&contract.SubmissionDate,
			&contract.ContractingBody,
			&contract.ScrapedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan contract: %w", err)
		}
		contracts = append(contracts, contract)
	}

	return contracts, nil
} 