package scraper

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

// CLIScraper handles web scraping using Selenium WebDriver in headless mode
type CLIScraper struct {
	driver      selenium.WebDriver
	coreScraper *CoreScraper
	sessionID   string // Unique session identifier for organizing screenshots
}

// NewCLIScraper creates a new CLI-only Selenium scraper instance (headless mode)
func NewCLIScraper() (*CLIScraper, error) {
	// Generate a unique session ID for this scraping session
	sessionID := fmt.Sprintf("cli_session_%s", time.Now().Format("2006-01-02_15-04-05"))
	
	// Chrome options for headless CLI operation
	chromeCaps := chrome.Capabilities{
		Args: []string{
			"--no-sandbox",
			"--disable-dev-shm-usage",
			"--headless",                    // Run in headless mode
			"--disable-gpu",                 // Disable GPU for headless
			"--window-size=1920,1080",       // Set window size for consistent rendering
			"--disable-web-security",        // Disable web security for scraping
			"--disable-features=VizDisplayCompositor", // Disable compositor for headless
			"--disable-extensions",          // Disable extensions for faster loading
			"--disable-plugins",             // Disable plugins
			"--disable-images",              // Disable images for faster loading
			"--disable-javascript-harmony-shipping", // Disable experimental JS features
		},
		W3C: true,
	}

	// Selenium capabilities
	caps := selenium.Capabilities{}
	caps.AddChrome(chromeCaps)
	
	// Add logging capabilities for CLI debugging
	caps["goog:loggingPrefs"] = map[string]string{
		"browser": "WARNING",  
		"driver":  "WARNING",
	}

	// Connect to Selenium server (trying both ports)
	var driver selenium.WebDriver
	var err error
	
	// Try port 4445 first, then 4446, then 4444
	for _, port := range []string{"4445", "4446", "4444"} {
		driver, err = selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%s", port))
		if err == nil {
			log.Printf("✅ Connected to ChromeDriver (CLI mode) on port %s", port)
			break
		}
		log.Printf("⚠️ Failed to connect to port %s: %v", port, err)
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to create CLI selenium driver on any port: %w", err)
	}

	// Test the headless browser
	if err := driver.Get("data:text/html,<html><body><h1>CLI Browser Test</h1></body></html>"); err == nil {
		log.Println("✅ CLI browser is responding to commands")
	} else {
		log.Printf("Warning: CLI browser test failed: %v", err)
	}

	return &CLIScraper{
		driver:      driver,
		coreScraper: NewCoreScraper(),
		sessionID:   sessionID,
	}, nil
}

// Close closes the CLI Selenium driver
func (c *CLIScraper) Close() error {
	if c.driver != nil {
		return c.driver.Quit()
	}
	return nil
}

// GetDriver returns the Selenium driver (for debugging purposes)
func (c *CLIScraper) GetDriver() selenium.WebDriver {
	return c.driver
}

// GetBaseURL returns the base URL
func (c *CLIScraper) GetBaseURL() string {
	return c.coreScraper.baseURL
}

// NavigateToSearchForm navigates to the search form page (CLI implementation)
func (c *CLIScraper) NavigateToSearchForm() error {
	log.Println("Step 1: Navigating directly to search form page (CLI mode)...")
	searchFormURL := c.coreScraper.GetSearchFormURL()
	
	if err := c.driver.Get(searchFormURL); err != nil {
		return fmt.Errorf("failed to navigate to search form page: %w", err)
	}

	log.Println("✅ Successfully navigated to search form page")
	log.Println("⏳ Waiting 8 seconds for page to fully load (CLI mode)...")
	time.Sleep(8 * time.Second) 

	// Take screenshot for debugging 
	if err := c.TakeScreenshotWithDescription("step1_search_form_navigation"); err != nil {
		log.Printf("Warning: Failed to take screenshot: %v", err)
	}

	// Debug the page structure to understand what's available
	log.Println("🔍 Debugging search form page structure (CLI mode)...")
	if err := c.DebugPageStructure(); err != nil {
		log.Printf("Warning: Page structure debugging failed: %v", err)
	}

	return nil
}

// EnterCPVCode enters the CPV code into the input field (CLI implementation)
func (c *CLIScraper) EnterCPVCode(code string) error {
	log.Println("Step 2: Setting CPV code (CLI mode)...")
	log.Println("🔍 Searching for CPV input field...")
	
	var cpvField selenium.WebElement
	
	// Try multiple selectors for CPV field (same as SeleniumScraper)
	selectors := []string{
		"//input[contains(@name, 'codigoCpv')]",
		"//input[contains(@name, 'cpv')]",
		"//input[contains(@id, 'cpv')]",
		"//input[contains(@id, 'codigo')]",
		"//input[@placeholder='CPV']",
		"//input[@placeholder='Código CPV']",
		"//input[@type='text' and contains(@class, 'form-control')]",
		"//input[@type='text' and contains(@class, 'input')]",
		"//input[@type='text' and contains(@style, 'width')]",
		"//input[@type='text']",
		"//input[contains(@class, 'form-control')]",
		"//input[contains(@class, 'input')]",
	}
	
	for _, selector := range selectors {
		log.Printf("🔍 Trying selector: %s", selector)
		var err error
		cpvField, err = c.driver.FindElement(selenium.ByXPATH, selector)
		if err == nil {
			log.Printf("✅ Found CPV field with selector: %s", selector)
			break
		}
	}
	
	if cpvField == nil {
		// If all selectors fail, try to get page source for debugging
		pageSource, _ := c.driver.PageSource()
		log.Printf("❌ Could not find CPV field. Page source preview: %s", pageSource[:500])
		return fmt.Errorf("could not find CPV input field")
	}

	log.Println("✅ Found CPV field, entering code...")
	log.Println("⏳ Clearing field and entering code in 2 seconds (CLI mode)...")
	time.Sleep(2 * time.Second) 
	
	// Clear and fill the CPV field
	if err := cpvField.Clear(); err != nil {
		return fmt.Errorf("failed to clear CPV field: %w", err)
	}
	
	// Type slowly to simulate human input (slightly faster for CLI mode)
	for _, char := range code {
		if err := cpvField.SendKeys(string(char)); err != nil {
			return fmt.Errorf("failed to enter CPV code: %w", err)
		}
		time.Sleep(50 * time.Millisecond) 
	}

	log.Println("✅ CPV code entered successfully")
	log.Println("⏳ Waiting 2 seconds (CLI mode)...")
	time.Sleep(2 * time.Second)

	// Take screenshot after entering CPV code
	if err := c.TakeScreenshotWithDescription("step2_cpv_code_entered"); err != nil {
		log.Printf("Warning: Failed to take screenshot: %v", err)
	}

	return nil
}

// ClickAnadirButton clicks the Añadir button (CLI implementation)
func (c *CLIScraper) ClickAnadirButton() error {
	log.Println("Step 3: Looking for 'Añadir' button (CLI mode)...")
	log.Println("🔍 Searching for Añadir button...")
	
	anadirButton, err := c.driver.FindElement(selenium.ByXPATH, "//input[@value='Añadir']")
	if err != nil {
		log.Printf("⚠️ Could not find Añadir button by value, trying alternative selectors...")
		
		// Try alternative selectors
		log.Println("🔍 Trying XPath: //button[contains(text(), 'Añadir')]")
		anadirButton, err = c.driver.FindElement(selenium.ByXPATH, "//button[contains(text(), 'Añadir')]")
		if err != nil {
			log.Println("🔍 Trying XPath: //input[@type='submit' and contains(@value, 'Añadir')]")
			anadirButton, err = c.driver.FindElement(selenium.ByXPATH, "//input[@type='submit' and contains(@value, 'Añadir')]")
			if err != nil {
				log.Println("🔍 Trying XPath: //*[contains(text(), 'Añadir')]")
				anadirButton, err = c.driver.FindElement(selenium.ByXPATH, "//*[contains(text(), 'Añadir')]")
				if err != nil {
					return fmt.Errorf("could not find Añadir button: %w", err)
				}
			}
		}
	}

	log.Println("✅ Found Añadir button, clicking...")
	log.Println("⏳ Clicking in 2 seconds (CLI mode)...")
	time.Sleep(2 * time.Second) 
	
	if err := anadirButton.Click(); err != nil {
		return fmt.Errorf("failed to click Añadir button: %w", err)
	}

	log.Println("✅ Successfully clicked Añadir button")
	log.Println("⏳ Waiting 3 seconds for form update (CLI mode)...")
	time.Sleep(3 * time.Second) 

	// Take screenshot after clicking Añadir
	if err := c.TakeScreenshotWithDescription("step3_anadir_button_clicked"); err != nil {
		log.Printf("Warning: Failed to take screenshot: %v", err)
	}

	return nil
}

// ClickBuscarButton clicks the Buscar button (CLI implementation)
func (c *CLIScraper) ClickBuscarButton() error {
	log.Println("Step 4: Looking for 'Buscar' button (CLI mode)...")
	log.Println("🔍 Searching for Buscar button...")
	
	buscarButton, err := c.driver.FindElement(selenium.ByXPATH, "//input[@value='Buscar']")
	if err != nil {
		log.Printf("⚠️ Could not find Buscar button by value, trying alternative selectors...")
		
		// Try alternative selectors
		log.Println("🔍 Trying XPath: //button[contains(text(), 'Buscar')]")
		buscarButton, err = c.driver.FindElement(selenium.ByXPATH, "//button[contains(text(), 'Buscar')]")
		if err != nil {
			log.Println("🔍 Trying XPath: //input[@type='submit']")
			buscarButton, err = c.driver.FindElement(selenium.ByXPATH, "//input[@type='submit']")
			if err != nil {
				log.Println("🔍 Trying XPath: //*[contains(text(), 'Buscar')]")
				buscarButton, err = c.driver.FindElement(selenium.ByXPATH, "//*[contains(text(), 'Buscar')]")
				if err != nil {
					return fmt.Errorf("could not find Buscar button: %w", err)
				}
			}
		}
	}

	log.Println("✅ Found Buscar button, clicking...")
	log.Println("⏳ Clicking in 2 seconds (CLI mode)...")
	time.Sleep(2 * time.Second) 
	
	if err := buscarButton.Click(); err != nil {
		return fmt.Errorf("failed to click Buscar button: %w", err)
	}

	log.Println("✅ Successfully clicked Buscar button")
	log.Println("⏳ Starting search process (CLI mode)...")

	return nil
}

// WaitForResults waits for the search results to load (CLI implementation)
func (c *CLIScraper) WaitForResults() error {
	log.Println("Step 5: Waiting for search results (CLI mode)...")
	
	// Wait for the loading to complete 
	maxWait := 45 * time.Second 
	startTime := time.Now()
	
	for time.Since(startTime) < maxWait {
		// Check if we're still on a loading page
		bodyText, err := c.driver.FindElement(selenium.ByTagName, "body")
		if err == nil {
			text, err := bodyText.Text()
			if err == nil {
				if strings.Contains(text, "Obteniendo búsqueda") || strings.Contains(text, "recuperando") {
					log.Println("⏳ Search still loading, waiting...")
					time.Sleep(3 * time.Second) 
					continue
				}
			}
		}
		
		// Check if results table is present
		_, err = c.driver.FindElement(selenium.ByID, "myTablaBusquedaCustom")
		if err == nil {
			log.Println("✅ Results table found!")
			break
		}
		
		log.Println("⏳ Still waiting for results table...")
		time.Sleep(2 * time.Second)
	}

	// Take screenshot after search
	if err := c.TakeScreenshotWithDescription("step4_search_results_loaded"); err != nil {
		log.Printf("Warning: Failed to take screenshot: %v", err)
	}

	return nil
}

// ExtractContracts extracts contracts from the results table (CLI implementation)
func (c *CLIScraper) ExtractContracts() ([]Contract, error) {
	log.Println("Step 6: Extracting contracts from results (CLI mode)...")
	
	// Get the page source (HTML content) from Selenium
	htmlContent, err := c.driver.PageSource()
	if err != nil {
		return nil, fmt.Errorf("failed to get page source: %w", err)
	}
	
	// Use the truly unified extraction method
	return c.coreScraper.ExtractContractsFromHTML(htmlContent)
}

// ExtractAllContracts extracts ALL contracts regardless of status for status change detection
func (c *CLIScraper) ExtractAllContracts() ([]Contract, error) {
	log.Println("Step 6b: Extracting ALL contracts for status change detection (CLI mode)...")
	
	// Get the page source (HTML content) from Selenium
	htmlContent, err := c.driver.PageSource()
	if err != nil {
		return nil, fmt.Errorf("failed to get page source: %w", err)
	}
	
	// Use the unified extraction method for all contracts
	return c.coreScraper.ExtractAllContractsFromHTML(htmlContent)
}



// GetSessionID returns the session ID
func (c *CLIScraper) GetSessionID() string {
	return c.sessionID
}

// TakeScreenshotWithDescription takes a screenshot with a descriptive name
func (c *CLIScraper) TakeScreenshotWithDescription(description string) error {
	// Create a clean filename from the description
	cleanDescription := strings.ReplaceAll(description, " ", "_")
	cleanDescription = strings.ReplaceAll(cleanDescription, "-", "_")
	cleanDescription = strings.ReplaceAll(cleanDescription, ".", "_")
	cleanDescription = strings.ReplaceAll(cleanDescription, ":", "_")
	
	filename := fmt.Sprintf("cli_%s.png", cleanDescription)
	return c.TakeScreenshot(filename)
}

// TakeScreenshot takes a screenshot for debugging (CLI mode)
func (c *CLIScraper) TakeScreenshot(filename string) error {
	bytes, err := c.driver.Screenshot()
	if err != nil {
		return fmt.Errorf("failed to take screenshot: %w", err)
	}

	// Create screenshots directory if it doesn't exist
	screenshotsDir := fmt.Sprintf("screenshots/%s", c.sessionID)
	if err := os.MkdirAll(screenshotsDir, 0755); err != nil {
		return fmt.Errorf("failed to create screenshots directory: %w", err)
	}

	// Generate timestamp for unique naming 
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	
	// Create a clean filename with timestamp
	cleanFilename := strings.ReplaceAll(filename, ".png", "")
	cleanFilename = strings.ReplaceAll(cleanFilename, " ", "_")
	cleanFilename = strings.ReplaceAll(cleanFilename, "-", "_")
	
	// Combine timestamp with clean filename
	timestampedFilename := fmt.Sprintf("%s_%s.png", timestamp, cleanFilename)
	
	// Full path for the screenshot
	fullPath := fmt.Sprintf("%s/%s", screenshotsDir, timestampedFilename)

	// Save screenshot to file
	if err := os.WriteFile(fullPath, bytes, 0644); err != nil {
		return fmt.Errorf("failed to save screenshot: %w", err)
	}

	log.Printf("📸 CLI Screenshot saved to: %s", fullPath)
	return nil
}

// DebugPageStructure analyzes and logs the page structure for debugging (CLI mode)
func (c *CLIScraper) DebugPageStructure() error {
	log.Println("🔍 Debugging page structure (CLI mode)...")
	
	// Get page title
	title, err := c.driver.Title()
	if err == nil {
		log.Printf("📄 Page title: %s", title)
	}
	
	// Get current URL
	currentURL, err := c.driver.CurrentURL()
	if err == nil {
		log.Printf("🌐 Current URL: %s", currentURL)
	}
	
	// Look for forms
	forms, err := c.driver.FindElements(selenium.ByTagName, "form")
	if err == nil {
		log.Printf("📝 Found %d forms on the page", len(forms))
		for i, form := range forms {
			action, _ := form.GetAttribute("action")
			method, _ := form.GetAttribute("method")
			log.Printf("  Form %d: action='%s', method='%s'", i+1, action, method)
		}
	}
	
	// Look for input fields
	inputs, err := c.driver.FindElements(selenium.ByTagName, "input")
	if err == nil {
		log.Printf("⌨️ Found %d input fields on the page", len(inputs))
		for i, input := range inputs {
			if i < 10 { // Limit to first 10 inputs to avoid spam
				name, _ := input.GetAttribute("name")
				id, _ := input.GetAttribute("id")
				value, _ := input.GetAttribute("value")
				inputType, _ := input.GetAttribute("type")
				log.Printf("  Input %d: name='%s', id='%s', type='%s', value='%s'", i+1, name, id, inputType, value)
			}
		}
		if len(inputs) > 10 {
			log.Printf("  ... and %d more inputs", len(inputs)-10)
		}
	}
	
	// Look for buttons
	buttons, err := c.driver.FindElements(selenium.ByTagName, "button")
	if err == nil {
		log.Printf("🔘 Found %d buttons on the page", len(buttons))
		for i, button := range buttons {
			if i < 5 { // Limit to first 5 buttons
				text, _ := button.Text()
				value, _ := button.GetAttribute("value")
				log.Printf("  Button %d: text='%s', value='%s'", i+1, text, value)
			}
		}
		if len(buttons) > 5 {
			log.Printf("  ... and %d more buttons", len(buttons)-5)
		}
	}
	
	// Look for tables
	tables, err := c.driver.FindElements(selenium.ByTagName, "table")
	if err == nil {
		log.Printf("📊 Found %d tables on the page", len(tables))
		for i, table := range tables {
			id, _ := table.GetAttribute("id")
			class, _ := table.GetAttribute("class")
			log.Printf("  Table %d: id='%s', class='%s'", i+1, id, class)
		}
	}
	
	log.Println("✅ Page structure debugging completed")
	return nil
}

// GetScreenshotsDirectory returns the screenshots directory path
func (c *CLIScraper) GetScreenshotsDirectory() string {
	return fmt.Sprintf("screenshots/%s", c.sessionID)
}

// ListScreenshots lists all screenshots taken in this session
func (c *CLIScraper) ListScreenshots() ([]string, error) {
	screenshotsDir := c.GetScreenshotsDirectory()
	
	// Check if directory exists
	if _, err := os.Stat(screenshotsDir); os.IsNotExist(err) {
		return []string{}, nil
	}
	
	// Read directory contents
	entries, err := os.ReadDir(screenshotsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read screenshots directory: %w", err)
	}
	
	var screenshots []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".png") {
			screenshots = append(screenshots, entry.Name())
		}
	}
	
	// Sort screenshots by name
	sort.Strings(screenshots)
	
	return screenshots, nil
}

// ExtractDocumentLinksFromContract visits a contract detail page and extracts Pliego and Anuncio links
func (c *CLIScraper) ExtractDocumentLinksFromContract(contractLink string) (pliegoLink, anuncioLink string, err error) {
	if contractLink == "" {
		return "", "", nil
	}
	
	log.Printf("🔍 Visiting contract detail page to extract document links...")
	
	// Navigate to the contract detail page
	if err := c.driver.Get(contractLink); err != nil {
		return "", "", fmt.Errorf("failed to navigate to contract detail page: %w", err)
	}
	
	// Wait for page to load
	time.Sleep(3 * time.Second)
	
	// Get the page source
	htmlContent, err := c.driver.PageSource()
	if err != nil {
		return "", "", fmt.Errorf("failed to get contract detail page source: %w", err)
	}
	
	// Extract document links using the core scraper method
	pliegoLink, anuncioLink = c.coreScraper.ExtractDocumentLinks(htmlContent)
	
	log.Printf("📄 Document links extracted - Pliego: %s, Anuncio: %s", 
		func() string { if pliegoLink != "" { return "✓" } else { return "✗" } }(),
		func() string { if anuncioLink != "" { return "✓" } else { return "✗" } }())
	
	return pliegoLink, anuncioLink, nil
}

// GetSessionInfo returns information about the current CLI session
func (c *CLIScraper) GetSessionInfo() map[string]interface{} {
	screenshots, _ := c.ListScreenshots()
	
	return map[string]interface{}{
		"session_id":     c.sessionID,
		"screenshots":    screenshots,
		"mode":           "CLI (Headless)",
		"base_url":       c.coreScraper.baseURL,
		"cpv_code":       c.coreScraper.cpvCode,
		"session_start":  time.Now().Format("2006-01-02 15:04:05"),
	}
} 