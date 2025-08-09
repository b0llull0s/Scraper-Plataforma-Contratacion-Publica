package scraper

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Contract represents a contract from the procurement platform
type Contract struct {
	ID                string    `json:"id"`
	Description       string    `json:"description"`
	ContractType      string    `json:"contract_type"`
	Status            string    `json:"status"`
	Amount            string    `json:"amount"`
	SubmissionDate    string    `json:"submission_date"`
	ContractingBody   string    `json:"contracting_body"`
	Link              string    `json:"link"`
	PliegoLink        string    `json:"pliego_link"`
	AnuncioLink       string    `json:"anuncio_link"`
	ScrapedAt         time.Time `json:"scraped_at"`
}

// ScraperInterface defines the interface that both HTTP and Selenium scrapers must implement
type ScraperInterface interface {
	NavigateToSearchForm() error
	EnterCPVCode(code string) error
	ClickAnadirButton() error
	ClickBuscarButton() error
	WaitForResults() error
	ExtractContracts() ([]Contract, error)
	ExtractAllContracts() ([]Contract, error)
	Close() error
}

// CoreScraper contains the unified business logic that orchestrates the scraping process
type CoreScraper struct {
	baseURL string
	cpvCode string
}

// NewCoreScraper creates a new core scraper with business logic
func NewCoreScraper() *CoreScraper {
	return &CoreScraper{
		baseURL: "https://contrataciondelestado.es",
		cpvCode: "32351200", // LED screens CPV code
	}
}

// GetSearchFormURL returns the direct URL to the search form
func (c *CoreScraper) GetSearchFormURL() string {
	return c.baseURL + "/wps/portal/!ut/p/b1/jdDLDoIwEAXQb-EDTKelFFiWZ0tQUAFtN6QLYzA8Nsbvtxq3orO7ybmZySCN1AYTHwcMh0DRGenZPIaruQ_LbMZX1qynaRXHmSAQHN0ESJm0LRM25p4FygLPjWlXdDU7yhxAiiwpW-xBTth_ffgyHH71T0ivE_IBaye-wcoNO7FMF6Qs83vepXsuQxeq6GAXFfW2qXOCwT6vQaqM0KTHLJQ3arjjPAFuDlpI/dl4/d5/L2dBISEvZ0FBIS9nQSEh/pw/Z7_AVEQAI930OBRD02JPMTPG21004/ren/p=sort_order=sortbiup/p=sort_id=sortHeaderEstado/p=_rvip=QCPjspQCPbusquedaQCPFormularioBusqueda.jsp/p=_rap=_rlnn/p=com.ibm.faces.portlet.mode=view/p=javax.servlet.include.path_info=QCPjspQCPbusquedaQCP_rlvid.jsp/-/#"
}

// GetCPVCode returns the CPV code to search for
func (c *CoreScraper) GetCPVCode() string {
	return c.cpvCode
}

// GetBaseURL returns the base URL
func (c *CoreScraper) GetBaseURL() string {
	return c.baseURL
}




// parseContractIDAndDescription separates the contract ID from the description
func (c *CoreScraper) parseContractIDAndDescription(fullText string) (id, description string) {
	fullText = strings.TrimSpace(fullText)
	
	// More comprehensive patterns for contract IDs
	patterns := []string{
		`^(\d{4,5}/\d{4})`,                    // Pattern: 10892/2024, 403/25
		`^(S-\d{5}-\d{4})`,                    // Pattern: S-02968-2025
		`^(\d{4}/\d{2})`,                      // Pattern: 2024/25
		`^([A-Z]-\d{5}-\d{4})`,                // Pattern: A-12345-2024
		`^(\d{4}-\d{2})`,                      // Pattern: 2024-25
		`^(\d{4}/[A-Z]+/\d{3}-\d{3}/\d{6})`,   // Pattern: 2025/D61000/006-201/00001
		`^([A-Z]+ CH SU-\d{2}-\d{2})`,         // Pattern: NGEU CH SU-02-25
		`^(\d{2}/\d{2})`,                      // Pattern: 13/25
		`^(\d{2}/\d{2}\.-[A-Z]+)`,             // Pattern: 13/25.-Suministro
		`^([A-Z]+\d{2}-\d{3}/\d{4})`,          // Pattern: 4AS25-815/2025
	}
	
	// Try exact pattern matches first
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if match := re.FindStringSubmatch(fullText); len(match) > 1 {
			id = match[1]
			description = strings.TrimSpace(fullText[len(id):])
			return
		}
	}
	
	// Look for the transition from ID to description
	// Common Spanish words that typically start contract descriptions
	descriptionStarters := []string{
		"Suministro", "Adquisición", "Contratación", "Servicios", "Instalación",
		"Alquiler", "Compra", "Adjudicación", "Ejecución", "Desarrollo",
		"Implementación", "Mantenimiento", "Reparación", "Renovación",
		"Ampliación", "Mejora", "Modernización", "Equipamiento", "Dotación",
	}
	
	// Try to find where the description starts
	for _, starter := range descriptionStarters {
		if idx := strings.Index(fullText, starter); idx > 0 {
			// Found a description starter, check if it's a reasonable split point
			potentialID := strings.TrimSpace(fullText[:idx])
			potentialDesc := strings.TrimSpace(fullText[idx:])
			
			// Validate that the potential ID looks like an ID (not too long, contains numbers/letters)
			if len(potentialID) > 0 && len(potentialID) <= 50 && 
			   (strings.ContainsAny(potentialID, "0123456789") || strings.Contains(potentialID, "/") || strings.Contains(potentialID, "-")) {
				id = potentialID
				description = potentialDesc
				return
			}
		}
	}
	
	// Fallback: Look for the first word that starts with a capital letter and is followed by lowercase
	// This is a more general approach
	for i := 1; i < len(fullText); i++ {
		if fullText[i] >= 'A' && fullText[i] <= 'Z' {
			// Found a capital letter, check if the previous character is not a letter/number
			// or if this looks like the start of a Spanish word
			if i > 0 && (fullText[i-1] < 'A' || fullText[i-1] > 'Z') && (fullText[i-1] < 'a' || fullText[i-1] > 'z') && (fullText[i-1] < '0' || fullText[i-1] > '9') {
				potentialID := strings.TrimSpace(fullText[:i])
				potentialDesc := strings.TrimSpace(fullText[i:])
				
				// Basic validation
				if len(potentialID) > 0 && len(potentialID) <= 50 {
					id = potentialID
					description = potentialDesc
					return
				}
			}
		}
	}
	
	// Last resort: if no clear pattern, use the first 30 characters as ID
	if len(fullText) > 30 {
		id = fullText[:30]
		description = fullText[30:]
	} else {
		id = fullText
		description = ""
	}
	
	return
}

// ScrapeLEDContracts is the unified main function that orchestrates the scraping process
// This is the single source of truth for the scraping workflow
func (c *CoreScraper) ScrapeLEDContracts(scraper ScraperInterface) ([]Contract, error) {
	log.Println("Starting LED contract scraper with unified logic...")
	
	// Step 1: Navigate to search form
	log.Println("Step 1: Navigating to search form...")
	if err := scraper.NavigateToSearchForm(); err != nil {
		return nil, fmt.Errorf("failed to navigate to search form: %w", err)
	}
	
	// Step 2: Enter CPV code
	log.Println("Step 2: Entering CPV code...")
	if err := scraper.EnterCPVCode(c.cpvCode); err != nil {
		return nil, fmt.Errorf("failed to enter CPV code: %w", err)
	}
	
	// Step 3: Click Añadir button
	log.Println("Step 3: Clicking Añadir button...")
	if err := scraper.ClickAnadirButton(); err != nil {
		return nil, fmt.Errorf("failed to click Añadir button: %w", err)
	}
	
	// Step 4: Click Buscar button
	log.Println("Step 4: Clicking Buscar button...")
	if err := scraper.ClickBuscarButton(); err != nil {
		return nil, fmt.Errorf("failed to click Buscar button: %w", err)
	}
	
	// Step 5: Wait for results
	log.Println("Step 5: Waiting for results...")
	if err := scraper.WaitForResults(); err != nil {
		return nil, fmt.Errorf("failed to wait for results: %w", err)
	}
	
	// Step 6: Extract contracts
	log.Println("Step 6: Extracting contracts...")
	contracts, err := scraper.ExtractContracts()
	if err != nil {
		return nil, fmt.Errorf("failed to extract contracts: %w", err)
	}
	
	log.Printf("Successfully extracted %d contracts with unified logic", len(contracts))
	return contracts, nil
}

// ExtractContractsFromTable is the unified method for extracting table data
// This method can be used by both HTTP and Selenium scrapers
func (c *CoreScraper) ExtractContractsFromTable(tableData [][]string) ([]Contract, error) {
	var contracts []Contract

	log.Printf("Processing %d rows of table data", len(tableData))

	// Process each row (skip header row if present)
	for i, row := range tableData {
		if i == 0 {
			// Check if this is a header row by looking for header-like content
			isHeader := false
			for _, cell := range row {
				lowerCell := strings.ToLower(strings.TrimSpace(cell))
				if strings.Contains(lowerCell, "expediente") || 
				   strings.Contains(lowerCell, "tipo") || 
				   strings.Contains(lowerCell, "estado") ||
				   strings.Contains(lowerCell, "importe") ||
				   strings.Contains(lowerCell, "presentación") ||
				   strings.Contains(lowerCell, "órgano") {
					isHeader = true
					break
				}
			}
			if isHeader {
				log.Println("Skipping header row")
				continue
			}
		}

		// Skip rows with insufficient cells
		if len(row) < 6 {
			log.Printf("Row %d has insufficient cells (%d), skipping", i, len(row))
			continue
		}

		// Parse the first column to separate ID and description
		id, description := c.parseContractIDAndDescription(row[0])
		
		// Extract contract data from row
		contract := Contract{
			ID:              id,
			Description:     description,
			ContractType:    strings.TrimSpace(row[1]),
			Status:          strings.TrimSpace(row[2]),
			Amount:          strings.TrimSpace(row[3]),
			SubmissionDate:  strings.TrimSpace(row[4]),
			ContractingBody: strings.TrimSpace(row[5]),
			ScrapedAt:       time.Now(),
		}

		// Only include NEW contracts with status "Publicada" (Published) or "Evaluación Previa" (Pre-evaluation)
		if strings.EqualFold(contract.Status, "Publicada") || strings.EqualFold(contract.Status, "Evaluación Previa") {
			contracts = append(contracts, contract)
			log.Printf("✅ Extracted contract (%s): %s", contract.Status, contract.ID)
		} else {
			log.Printf("⏭️ Skipped contract (status: %s): %s", contract.Status, contract.ID)
		}
	}

	log.Printf("Extracted %d contracts from table data", len(contracts))
	return contracts, nil
}

// ExtractContractsFromTableWithLinks extracts contracts from table data with links
func (c *CoreScraper) ExtractContractsFromTableWithLinks(tableData [][]string, links []string) ([]Contract, error) {
	var contracts []Contract

	log.Printf("Processing %d rows of table data with links", len(tableData))

	// Process each row (skip header row if present)
	for i, row := range tableData {
		if i == 0 {
			// Check if this is a header row by looking for header-like content
			isHeader := false
			for _, cell := range row {
				lowerCell := strings.ToLower(strings.TrimSpace(cell))
				if strings.Contains(lowerCell, "expediente") || 
				   strings.Contains(lowerCell, "tipo") || 
				   strings.Contains(lowerCell, "estado") ||
				   strings.Contains(lowerCell, "importe") ||
				   strings.Contains(lowerCell, "presentación") ||
				   strings.Contains(lowerCell, "órgano") {
					isHeader = true
					break
				}
			}
			if isHeader {
				log.Println("Skipping header row")
				continue
			}
		}

		// Skip rows with insufficient cells
		if len(row) < 6 {
			log.Printf("Row %d has insufficient cells (%d), skipping", i, len(row))
			continue
		}

		// Parse the first column to separate ID and description
		id, description := c.parseContractIDAndDescription(row[0])
		
		// Get the link for this contract (if available)
		link := ""
		if i < len(links) {
			link = links[i]
		}
		
		// Try to extract document links from the current row if available
		pliegoLink, anuncioLink := c.extractDocumentLinksFromRow(row)
		
		// Extract contract data from row
		contract := Contract{
			ID:              id,
			Description:     description,
			ContractType:    strings.TrimSpace(row[1]),
			Status:          strings.TrimSpace(row[2]),
			Amount:          strings.TrimSpace(row[3]),
			SubmissionDate:  strings.TrimSpace(row[4]),
			ContractingBody: strings.TrimSpace(row[5]),
			Link:            link,
			PliegoLink:      pliegoLink,
			AnuncioLink:     anuncioLink,
			ScrapedAt:       time.Now(),
		}

		// Only include NEW contracts with status "Publicada" (Published) or "Evaluación Previa" (Pre-evaluation)
		if strings.EqualFold(contract.Status, "Publicada") || strings.EqualFold(contract.Status, "Evaluación Previa") {
			contracts = append(contracts, contract)
			log.Printf("✅ Extracted contract (%s): %s", contract.Status, contract.ID)
		} else {
			log.Printf("⏭️ Skipped contract (status: %s): %s", contract.Status, contract.ID)
		}
	}

	log.Printf("Extracted %d contracts from table data with links", len(contracts))
	return contracts, nil
}

// ExtractDocumentLinks extracts Pliego and Anuncio de Licitación links from a contract detail page
func (c *CoreScraper) ExtractDocumentLinks(htmlContent string) (pliegoLink, anuncioLink string) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		log.Printf("Failed to parse contract detail HTML: %v", err)
		return "", ""
	}

	// Debug: Log the page structure
	log.Printf("🔍 Analyzing contract detail page structure...")
	
	// Count all links on the page
	allLinks := doc.Find("a")
	log.Printf("📊 Found %d total links on the contract detail page", allLinks.Length())
	
	// Look for links with class "celdaTam2" that contain the document links
	celdaTam2Links := doc.Find("a.celdaTam2")
	log.Printf("📊 Found %d links with class 'celdaTam2'", celdaTam2Links.Length())
	
	// Look for any links containing GetDocumentByIdServlet
	documentLinks := doc.Find("a[href*='GetDocumentByIdServlet']")
	log.Printf("📊 Found %d links containing 'GetDocumentByIdServlet'", documentLinks.Length())
	
	// Log all document links for debugging
	documentLinks.Each(func(i int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		text := strings.TrimSpace(s.Text())
		parentText := s.Parent().Text()
		// Safely truncate parent text to avoid slice bounds error
		parentPreview := parentText
		if len(parentText) > 100 {
			parentPreview = parentText[:100]
		}
		log.Printf("🔗 Document link %d: href='%s', text='%s', parent='%s'", i+1, href, text, parentPreview)
	})

	// Look for links with class "celdaTam2" that contain the document links
	doc.Find("a.celdaTam2").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists {
			return
		}

		// Check if this is a document link (contains GetDocumentByIdServlet)
		if strings.Contains(href, "GetDocumentByIdServlet") {
			// Find the document type by looking at the table row structure
			// The document type is in the second column (td.tipoDocumento) of the same row
			row := s.Closest("tr")
			if row.Length() > 0 {
				documentTypeCell := row.Find("td.tipoDocumento")
				if documentTypeCell.Length() > 0 {
					documentType := strings.TrimSpace(documentTypeCell.Text())
					lowerDocumentType := strings.ToLower(documentType)
					
					log.Printf("🔍 Found document link with type: '%s'", documentType)
					
					// Look for Pliego link
					if strings.Contains(lowerDocumentType, "pliego") {
						pliegoLink = href
						log.Printf("🔗 Found Pliego link: %s", href)
					}
					
					// Look for Anuncio de Licitación link
					if strings.Contains(lowerDocumentType, "anuncio") || 
					   strings.Contains(lowerDocumentType, "licitación") ||
					   strings.Contains(lowerDocumentType, "rectificación") {
						anuncioLink = href
						log.Printf("🔗 Found Anuncio de Licitación link: %s", href)
					}
				}
			}
		}
	})

	return pliegoLink, anuncioLink
}

// extractDocumentLinksFromRow attempts to extract document links from a table row
// This is a fallback method in case document links are embedded in the search results
func (c *CoreScraper) extractDocumentLinksFromRow(row []string) (pliegoLink, anuncioLink string) {
	// For now, this is a placeholder. The document links are typically not in the search results table
	// but rather on the individual contract detail pages
	return "", ""
}

// EnhanceContractsWithDocumentLinks visits each contract detail page and extracts document links
// This method requires a Selenium scraper to navigate to individual contract pages
// It also accepts a storage interface to check if contracts already have document links
func (c *CoreScraper) EnhanceContractsWithDocumentLinks(contracts []Contract, seleniumScraper interface{}, storage interface{}) ([]Contract, error) {
	enhancedContracts := make([]Contract, len(contracts))
	
	log.Printf("🔍 Starting document link enhancement for %d contracts...", len(contracts))
	
	// Count contracts that will be processed vs skipped
	contractsToProcess := 0
	contractsToSkip := 0
	
	for i, contract := range contracts {
		enhancedContracts[i] = contract
		
		// Skip if no contract link available
		if contract.Link == "" {
			log.Printf("⚠️ No contract link available for %s, skipping document extraction", contract.ID)
			contractsToSkip++
			continue
		}
		
		// Check if contract already has document links in the database
		if storage != nil {
			// Try to cast to the interface
			storageInterface, ok := storage.(interface {
				GetContractByID(string) (*Contract, error)
			})
			
			if ok {
				existingContract, err := storageInterface.GetContractByID(contract.ID)
				if err != nil {
					log.Printf("⚠️ Failed to check existing contract %s: %v", contract.ID, err)
				} else if existingContract != nil {
					if existingContract.PliegoLink != "" && existingContract.AnuncioLink != "" {
						// Contract already has both document links, skip extraction
						log.Printf("⏭️ Contract %s already has document links, skipping extraction", contract.ID)
						enhancedContracts[i].PliegoLink = existingContract.PliegoLink
						enhancedContracts[i].AnuncioLink = existingContract.AnuncioLink
						contractsToSkip++
						continue
					} else if existingContract.PliegoLink != "" || existingContract.AnuncioLink != "" {
						// Contract has partial document links, we'll try to complete them
						log.Printf("🔄 Contract %s has partial document links, attempting to complete...", contract.ID)
						enhancedContracts[i].PliegoLink = existingContract.PliegoLink
						enhancedContracts[i].AnuncioLink = existingContract.AnuncioLink
					}
				}
			}
		}
		
		log.Printf("🔍 Processing contract %s with link: %s", contract.ID, contract.Link)
		contractsToProcess++
		
		// Try to extract document links using Selenium scraper
		if scraper, ok := seleniumScraper.(interface {
			ExtractDocumentLinksFromContract(string) (string, string, error)
		}); ok {
			log.Printf("✅ Found compatible scraper, extracting document links for %s...", contract.ID)
			pliegoLink, anuncioLink, err := scraper.ExtractDocumentLinksFromContract(contract.Link)
			if err != nil {
				log.Printf("⚠️ Failed to extract document links for contract %s: %v", contract.ID, err)
				continue
			}
			
			// Only update if we got new links (don't overwrite existing ones with empty values)
			if pliegoLink != "" {
				enhancedContracts[i].PliegoLink = pliegoLink
			}
			if anuncioLink != "" {
				enhancedContracts[i].AnuncioLink = anuncioLink
			}
			
			log.Printf("📄 Enhanced contract %s with document links - Pliego: %s, Anuncio: %s", 
				contract.ID, 
				func() string { if enhancedContracts[i].PliegoLink != "" { return "✓" } else { return "✗" } }(),
				func() string { if enhancedContracts[i].AnuncioLink != "" { return "✓" } else { return "✗" } }())
		} else {
			log.Printf("❌ Selenium scraper does not implement ExtractDocumentLinksFromContract method")
		}
	}
	
	log.Printf("✅ Document link enhancement completed - Processed: %d, Skipped: %d", contractsToProcess, contractsToSkip)
	return enhancedContracts, nil
}

// ExtractAllContractsFromTable extracts ALL contracts regardless of status for status change detection
func (c *CoreScraper) ExtractAllContractsFromTable(tableData [][]string) ([]Contract, error) {
	var allContracts []Contract

	log.Printf("Processing %d rows for status change detection", len(tableData))

	// Process each row (skip header row if present)
	for i, row := range tableData {
		if i == 0 {
			// Check if this is a header row by looking for header-like content
			isHeader := false
			for _, cell := range row {
				lowerCell := strings.ToLower(strings.TrimSpace(cell))
				if strings.Contains(lowerCell, "expediente") || 
				   strings.Contains(lowerCell, "tipo") || 
				   strings.Contains(lowerCell, "estado") ||
				   strings.Contains(lowerCell, "importe") ||
				   strings.Contains(lowerCell, "presentación") ||
				   strings.Contains(lowerCell, "órgano") {
					isHeader = true
					break
				}
			}
			if isHeader {
				log.Println("Skipping header row")
				continue
			}
		}

		// Skip rows with insufficient cells
		if len(row) < 6 {
			log.Printf("Row %d has insufficient cells (%d), skipping", i, len(row))
			continue
		}

		// Parse the first column to separate ID and description
		id, description := c.parseContractIDAndDescription(row[0])
		
		// Extract contract data from row
		contract := Contract{
			ID:              id,
			Description:     description,
			ContractType:    strings.TrimSpace(row[1]),
			Status:          strings.TrimSpace(row[2]),
			Amount:          strings.TrimSpace(row[3]),
			SubmissionDate:  strings.TrimSpace(row[4]),
			ContractingBody: strings.TrimSpace(row[5]),
			ScrapedAt:       time.Now(),
		}

		// Include ALL contracts for status change detection
		allContracts = append(allContracts, contract)
		log.Printf("📋 Found contract (%s): %s", contract.Status, contract.ID)
	}

	log.Printf("Found %d contracts for status change detection", len(allContracts))
	return allContracts, nil
}

// ExtractContractsFromHTML is the truly unified method that both HTTP and Selenium can use
// This method takes raw HTML and extracts table data using the same logic
func (c *CoreScraper) ExtractContractsFromHTML(htmlContent string) ([]Contract, error) {
	// Parse HTML using goquery (same for both HTTP and Selenium)
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Find the results table - EXACTLY the same for both
	table := doc.Find("#myTablaBusquedaCustom")
	if table.Length() == 0 {
		return nil, fmt.Errorf("could not find results table")
	}

	// Get all rows in the table - EXACTLY the same for both
	rows := table.Find("tr")
	log.Printf("Found %d rows in results table", rows.Length())

	// Convert table data to string matrix and extract links - EXACTLY the same for both
	var tableData [][]string
	var links []string
	
	rows.Each(func(i int, row *goquery.Selection) {
		// Get cells in this row - EXACTLY the same for both
		cells := row.Find("td")
		
		// Convert cells to string array - EXACTLY the same for both
		var rowData []string
		var link string
		
		cells.Each(func(j int, cell *goquery.Selection) {
			text := strings.TrimSpace(cell.Text())
			rowData = append(rowData, text)
			
			// Extract link from the first cell (contract ID cell)
			if j == 0 {
				// Look specifically for the contract detail link (the one with detalle_licitacion)
				linkElement := cell.Find("a[href*='detalle_licitacion']")
				if linkElement.Length() > 0 {
					if href, exists := linkElement.Attr("href"); exists {
						// This is the proper contract detail URL - use it directly
						link = href
						log.Printf("🔗 Found contract detail link: %s", href)
					}
				} else {
					// Fallback: look for any other link
					linkElement := cell.Find("a")
					if linkElement.Length() > 0 {
						if href, exists := linkElement.Attr("href"); exists {
							// Convert relative links to absolute URLs
							if strings.HasPrefix(href, "#") {
								// This is a JavaScript link, provide a generic search URL
								link = c.baseURL + "/wps/portal/!ut/p/b1/jdDLDoIwEAXQb-EDTKelFFiWZ0tQUAFtN6QLYzA8Nsbvtxq3orO7ybmZySCN1AYTHwcMh0DRGenZPIaruQ_LbMZX1qynaRXHmSAQHN0ESJm0LRM25p4FygLPjWlXdDU7yhxAiiwpW-xBTth_ffgyHH71T0ivE_IBaye-wcoNO7FMF6Qs83vepXsuQxeq6GAXFfW2qXOCwT6vQaqM0KTHLJQ3arjjPAFuDlpI/dl4/d5/L2dBISEvZ0FBIS9nQSEh/pw/Z7_AVEQAI930OBRD02JPMTPG21004/ren/p=sort_order=sortbiup/p=sort_id=sortHeaderEstado/p=_rvip=QCPjspQCPbusquedaQCPFormularioBusqueda.jsp/p=_rap=_rlnn/p=com.ibm.faces.portlet.mode=view/p=javax.servlet.include.path_info=QCPjspQCPbusquedaQCP_rlvid.jsp/-/#"
							} else if strings.HasPrefix(href, "/") {
								// Relative URL starting with /
								link = c.baseURL + href
							} else if strings.HasPrefix(href, "https://contrataciondelestado.es/wps/poc") {
								// This is the proper contract detail URL
								link = href
							} else if !strings.HasPrefix(href, "http") {
								// Relative URL without /
								link = c.baseURL + "/" + href
							} else {
								// Already absolute URL
								link = href
							}
						}
					}
				}
			}
		})
		
		// Only add rows with sufficient data - EXACTLY the same for both
		if len(rowData) >= 6 {
			tableData = append(tableData, rowData)
			links = append(links, link)
		} else {
			log.Printf("Row %d has insufficient cells (%d), skipping", i, len(rowData))
		}
	})

	// Use the unified extraction logic from CoreScraper with links
	return c.ExtractContractsFromTableWithLinks(tableData, links)
}

// ExtractAllContractsFromHTML extracts ALL contracts regardless of status for status change detection
func (c *CoreScraper) ExtractAllContractsFromHTML(htmlContent string) ([]Contract, error) {
	// Parse HTML using goquery (same for both HTTP and Selenium)
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Find the results table - EXACTLY the same for both
	table := doc.Find("#myTablaBusquedaCustom")
	if table.Length() == 0 {
		return nil, fmt.Errorf("could not find results table")
	}

	// Get all rows in the table - EXACTLY the same for both
	rows := table.Find("tr")
	log.Printf("Found %d rows in results table for status change detection", rows.Length())

	// Convert table data to string matrix - EXACTLY the same for both
	var tableData [][]string
	
	rows.Each(func(i int, row *goquery.Selection) {
		// Get cells in this row - EXACTLY the same for both
		cells := row.Find("td")
		
		// Convert cells to string array - EXACTLY the same for both
		var rowData []string
		cells.Each(func(j int, cell *goquery.Selection) {
			text := strings.TrimSpace(cell.Text())
			rowData = append(rowData, text)
		})
		
		// Only add rows with sufficient data - EXACTLY the same for both
		if len(rowData) >= 6 {
			tableData = append(tableData, rowData)
		} else {
			log.Printf("Row %d has insufficient cells (%d), skipping", i, len(rowData))
		}
	})

	// Use the unified extraction logic for all contracts
	return c.ExtractAllContractsFromTable(tableData)
}






// ScraperType defines the type of scraper to use
type ScraperType string

const (
	ScraperTypeSelenium ScraperType = "selenium"
	ScraperTypeCLI      ScraperType = "cli"
)

// NewScraper creates a new scraper based on the specified type
func NewScraper(scraperType ScraperType) (ScraperInterface, error) {
	switch scraperType {
	case ScraperTypeSelenium:
		return NewSeleniumScraper()
	case ScraperTypeCLI:
		return NewCLIScraper()
	default:
		return nil, fmt.Errorf("unknown scraper type: %s", scraperType)
	}
}

// ScrapeContracts is the unified function that works with any scraper type
func ScrapeContracts(scraperType ScraperType) ([]Contract, error) {
	scraper, err := NewScraper(scraperType)
	if err != nil {
		return nil, fmt.Errorf("failed to create scraper: %w", err)
	}
	defer scraper.Close()

	coreScraper := NewCoreScraper()
	return coreScraper.ScrapeLEDContracts(scraper)
}

// ScrapeContractsWithScraper is a helper function that works with a specific scraper instance
func ScrapeContractsWithScraper(scraper ScraperInterface) ([]Contract, error) {
	coreScraper := NewCoreScraper()
	return coreScraper.ScrapeLEDContracts(scraper)
}

 