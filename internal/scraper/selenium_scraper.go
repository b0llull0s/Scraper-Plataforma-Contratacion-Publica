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

// SeleniumScraper handles web scraping using Selenium WebDriver
type SeleniumScraper struct {
	driver      selenium.WebDriver
	coreScraper *CoreScraper
	sessionID   string 
}

// NewSeleniumScraper creates a new Selenium scraper instance
func NewSeleniumScraper() (*SeleniumScraper, error) {
	// Generate a unique session ID for this scraping session
	sessionID := fmt.Sprintf("session_%s", time.Now().Format("2006-01-02_15-04-05"))
	
	// Chrome options for visible browser (simple and direct)
	chromeCaps := chrome.Capabilities{
		Args: []string{
			"--no-sandbox",
			"--disable-dev-shm-usage",
			"--window-size=1200,800",
			"--start-maximized",
		},
		W3C: true,
	}

	// Selenium capabilities
	caps := selenium.Capabilities{}
	caps.AddChrome(chromeCaps)
	
	// Add logging capabilities
	caps["goog:loggingPrefs"] = map[string]string{
		"browser": "ALL",
		"driver":  "ALL",
	}

	// Connect to Selenium server (trying both ports)
	var driver selenium.WebDriver
	var err error
	
	// Try port 4445 first, then 4446, then 4444
	for _, port := range []string{"4445", "4446", "4444"} {
		driver, err = selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%s", port))
		if err == nil {
			log.Printf("✅ Connected to ChromeDriver on port %s", port)
			break
		}
		log.Printf("⚠️ Failed to connect to port %s: %v", port, err)
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to create selenium driver on any port: %w", err)
	}

	// Set window size to be visible
	if err := driver.ResizeWindow("", 1920, 1080); err != nil {
		log.Printf("Warning: Could not resize window: %v", err)
	}

	// Bring window to front
	if err := driver.MaximizeWindow(""); err != nil {
		log.Printf("Warning: Could not maximize window: %v", err)
	}

	// Take a screenshot immediately to verify browser is working
	if err := driver.Get("data:text/html,<html><body><h1>Browser Test</h1></body></html>"); err == nil {
		log.Println("✅ Browser is responding to commands")
	} else {
		log.Printf("Warning: Browser test failed: %v", err)
	}

	return &SeleniumScraper{
		driver:      driver,
		coreScraper: NewCoreScraper(),
		sessionID:   sessionID,
	}, nil
}

// Close closes the Selenium driver
func (s *SeleniumScraper) Close() error {
	if s.driver != nil {
		return s.driver.Quit()
	}
	return nil
}

// GetDriver returns the Selenium driver (for debugging purposes)
func (s *SeleniumScraper) GetDriver() selenium.WebDriver {
	return s.driver
}

// GetBaseURL returns the base URL
func (s *SeleniumScraper) GetBaseURL() string {
	return s.coreScraper.baseURL
}

// NavigateToSearchForm navigates to the search form page
func (s *SeleniumScraper) NavigateToSearchForm() error {
	log.Println("Step 1: Navigating directly to search form page...")
	searchFormURL := s.coreScraper.GetSearchFormURL()
	
	if err := s.driver.Get(searchFormURL); err != nil {
		return fmt.Errorf("failed to navigate to search form page: %w", err)
	}

	log.Println("✅ Successfully navigated to search form page")
	log.Println("⏳ Waiting 10 seconds for page to fully load...")
	time.Sleep(10 * time.Second)

	// Take screenshot after navigation
	if err := s.TakeScreenshotWithDescription("step1_search_form_navigation"); err != nil {
		log.Printf("Warning: Failed to take screenshot: %v", err)
	}

	// Debug the page structure to understand what's available
	log.Println("🔍 Debugging search form page structure...")
	if err := s.DebugPageStructure(); err != nil {
		log.Printf("Warning: Page structure debugging failed: %v", err)
	}

	return nil
}

// EnterCPVCode enters the CPV code into the input field
func (s *SeleniumScraper) EnterCPVCode(code string) error {
	log.Println("Step 2: Setting CPV code...")
	log.Println("🔍 Searching for CPV input field...")
	
	var cpvField selenium.WebElement
	
	// Try multiple selectors for CPV field
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
		cpvField, err = s.driver.FindElement(selenium.ByXPATH, selector)
		if err == nil {
			log.Printf("✅ Found CPV field with selector: %s", selector)
			break
		}
	}
	
	if cpvField == nil {
		// If all selectors fail, try to get page source for debugging
		pageSource, _ := s.driver.PageSource()
		log.Printf("❌ Could not find CPV field. Page source preview: %s", pageSource[:500])
		return fmt.Errorf("could not find CPV input field")
	}

	log.Println("✅ Found CPV field, entering code...")
	log.Println("⏳ Clearing field and entering code in 3 seconds...")
	time.Sleep(3 * time.Second)
	
	// Clear and fill the CPV field
	if err := cpvField.Clear(); err != nil {
		return fmt.Errorf("failed to clear CPV field: %w", err)
	}
	
	// Type slowly to simulate human input
	for _, char := range code {
		if err := cpvField.SendKeys(string(char)); err != nil {
			return fmt.Errorf("failed to enter CPV code: %w", err)
		}
		time.Sleep(100 * time.Millisecond) // Type like a human
	}

	log.Println("✅ CPV code entered successfully")
	log.Println("⏳ Waiting 3 seconds...")
	time.Sleep(3 * time.Second)

	// Take screenshot after entering CPV
	if err := s.TakeScreenshotWithDescription("step2_cpv_code_entered"); err != nil {
		log.Printf("Warning: Failed to take screenshot: %v", err)
	}

	return nil
}


// ClickAnadirButton clicks the "Añadir" button
func (s *SeleniumScraper) ClickAnadirButton() error {
	log.Println("Step 3: Looking for 'Añadir' button...")
	log.Println("🔍 Searching for Añadir button...")
	
	anadirButton, err := s.driver.FindElement(selenium.ByXPATH, "//input[@value='Añadir']")
	if err != nil {
		log.Printf("⚠️ Could not find Añadir button by value, trying alternative selectors...")
		
		// Try alternative selectors
		log.Println("🔍 Trying XPath: //a[contains(text(), 'Añadir')]")
		anadirButton, err = s.driver.FindElement(selenium.ByXPATH, "//a[contains(text(), 'Añadir')]")
		if err != nil {
			log.Println("🔍 Trying XPath: //span[contains(text(), 'Añadir')]")
			anadirButton, err = s.driver.FindElement(selenium.ByXPATH, "//span[contains(text(), 'Añadir')]")
			if err != nil {
				log.Println("🔍 Trying XPath: //button[contains(text(), 'Añadir')]")
				anadirButton, err = s.driver.FindElement(selenium.ByXPATH, "//button[contains(text(), 'Añadir')]")
				if err != nil {
					log.Println("🔍 Trying XPath: //*[contains(text(), 'Añadir')]")
					anadirButton, err = s.driver.FindElement(selenium.ByXPATH, "//*[contains(text(), 'Añadir')]")
					if err != nil {
						return fmt.Errorf("could not find Añadir button: %w", err)
					}
				}
			}
		}
	}

	log.Println("✅ Found Añadir button, clicking...")
	log.Println("⏳ Clicking in 3 seconds...")
	time.Sleep(3 * time.Second)
	
	if err := anadirButton.Click(); err != nil {
		return fmt.Errorf("failed to click Añadir button: %w", err)
	}

	log.Println("✅ Successfully clicked Añadir button")
	log.Println("⏳ Waiting 5 seconds for the CPV to be added...")
	time.Sleep(5 * time.Second)

	// Take screenshot after clicking Añadir
	if err := s.TakeScreenshotWithDescription("step3_anadir_button_clicked"); err != nil {
		log.Printf("Warning: Failed to take screenshot: %v", err)
	}

	return nil
}

// ClickBuscarButton clicks the "Buscar" button
func (s *SeleniumScraper) ClickBuscarButton() error {
	log.Println("Step 4: Looking for 'Buscar' button...")
	log.Println("🔍 Searching for Buscar button...")
	
	buscarButton, err := s.driver.FindElement(selenium.ByXPATH, "//input[@value='Buscar']")
	if err != nil {
		log.Printf("⚠️ Could not find Buscar button by value, trying alternative selectors...")
		
		// Try alternative selectors
		log.Println("🔍 Trying XPath: //button[contains(text(), 'Buscar')]")
		buscarButton, err = s.driver.FindElement(selenium.ByXPATH, "//button[contains(text(), 'Buscar')]")
		if err != nil {
			log.Println("🔍 Trying XPath: //input[@type='submit']")
			buscarButton, err = s.driver.FindElement(selenium.ByXPATH, "//input[@type='submit']")
			if err != nil {
				log.Println("🔍 Trying XPath: //*[contains(text(), 'Buscar')]")
				buscarButton, err = s.driver.FindElement(selenium.ByXPATH, "//*[contains(text(), 'Buscar')]")
				if err != nil {
					return fmt.Errorf("could not find Buscar button: %w", err)
				}
			}
		}
	}

	log.Println("✅ Found Buscar button, clicking...")
	log.Println("⏳ Clicking in 3 seconds...")
	time.Sleep(3 * time.Second)
	
	if err := buscarButton.Click(); err != nil {
		return fmt.Errorf("failed to click Buscar button: %w", err)
	}

	log.Println("✅ Successfully clicked Buscar button")
	log.Println("⏳ Starting search process...")

	return nil
}

// WaitForResults waits for the search results to load
func (s *SeleniumScraper) WaitForResults() error {
	log.Println("Step 5: Waiting for search results...")
	
	// Wait for the loading to complete
	maxWait := 60 * time.Second
	startTime := time.Now()
	
	for time.Since(startTime) < maxWait {
		// Check if we're still on a loading page
		bodyText, err := s.driver.FindElement(selenium.ByTagName, "body")
		if err == nil {
			text, err := bodyText.Text()
			if err == nil {
				if strings.Contains(text, "Obteniendo búsqueda") || strings.Contains(text, "recuperando") {
					log.Println("⏳ Search still loading, waiting...")
					time.Sleep(5 * time.Second)
					continue
				}
			}
		}
		
		// Check if results table is present
		_, err = s.driver.FindElement(selenium.ByID, "myTablaBusquedaCustom")
		if err == nil {
			log.Println("✅ Results table found!")
			break
		}
		
		log.Println("⏳ Still waiting for results table...")
		time.Sleep(2 * time.Second)
	}

	// Take screenshot after search
	if err := s.TakeScreenshotWithDescription("step4_search_results_loaded"); err != nil {
		log.Printf("Warning: Failed to take screenshot: %v", err)
	}

	return nil
}

// ExtractContracts extracts contracts from the results table
func (s *SeleniumScraper) ExtractContracts() ([]Contract, error) {
	log.Println("Step 6: Extracting contracts from results...")
	
	// Get the page source (HTML content) from Selenium
	htmlContent, err := s.driver.PageSource()
	if err != nil {
		return nil, fmt.Errorf("failed to get page source: %w", err)
	}
	
	// Use the truly unified extraction method
	return s.coreScraper.ExtractContractsFromHTML(htmlContent)
}

// ExtractAllContracts extracts ALL contracts regardless of status for status change detection
func (s *SeleniumScraper) ExtractAllContracts() ([]Contract, error) {
	log.Println("Step 6b: Extracting ALL contracts for status change detection...")
	
	// Get the page source (HTML content) from Selenium
	htmlContent, err := s.driver.PageSource()
	if err != nil {
		return nil, fmt.Errorf("failed to get page source: %w", err)
	}
	
	// Use the unified extraction method for all contracts
	return s.coreScraper.ExtractAllContractsFromHTML(htmlContent)
}

// ExtractDocumentLinksFromContract visits a contract detail page and extracts Pliego and Anuncio links
func (s *SeleniumScraper) ExtractDocumentLinksFromContract(contractLink string) (pliegoLink, anuncioLink string, err error) {
	if contractLink == "" {
		return "", "", nil
	}
	
	log.Printf("🔍 Visiting contract detail page to extract document links...")
	
	// Navigate to the contract detail page
	if err := s.driver.Get(contractLink); err != nil {
		return "", "", fmt.Errorf("failed to navigate to contract detail page: %w", err)
	}
	
	// Wait for page to load
	time.Sleep(3 * time.Second)
	
	// Get the page source
	htmlContent, err := s.driver.PageSource()
	if err != nil {
		return "", "", fmt.Errorf("failed to get contract detail page source: %w", err)
	}
	
	// Extract document links using the core scraper method
	pliegoLink, anuncioLink = s.coreScraper.ExtractDocumentLinks(htmlContent)
	
	log.Printf("📄 Document links extracted - Pliego: %s, Anuncio: %s", 
		func() string { if pliegoLink != "" { return "✓" } else { return "✗" } }(),
		func() string { if anuncioLink != "" { return "✓" } else { return "✗" } }())
	
	return pliegoLink, anuncioLink, nil
}





// FindLicitacionesLink finds the Licitaciones link using multiple strategies
func (s *SeleniumScraper) FindLicitacionesLink() (selenium.WebElement, error) {
	log.Println("🔍 Looking for Licitaciones link with multiple strategies...")
	
	// Strategy 1: Try the original ID
	log.Println("Strategy 1: Trying original ID...")
	licitacionesLink, err := s.driver.FindElement(selenium.ByID, "viewns_Z7_AVEQAI930OBRD02JPMTPG21004_:form1:linkFormularioBusqueda")
	if err == nil {
		log.Println("✅ Found Licitaciones link by original ID")
		return licitacionesLink, nil
	}
	
	// Strategy 2: Try XPath with text content
	log.Println("Strategy 2: Trying XPath with text content...")
	selectors := []string{
		"//a[contains(text(), 'Licitaciones')]",
		"//a[contains(text(), 'Búsqueda de licitaciones')]",
		"//a[contains(text(), 'formulario')]",
		"//a[contains(text(), 'busqueda')]",
		"//a[contains(@href, 'formulario')]",
		"//a[contains(@href, 'busqueda')]",
		"//a[contains(@href, 'licitaciones')]",
		"//a[contains(@class, 'link')]",
		"//a[contains(@class, 'button')]",
		"//a[contains(@class, 'btn')]",
		"//span[contains(text(), 'Búsqueda de licitaciones por formulario')]/parent::a",
		"//span[contains(text(), 'Búsqueda de licitaciones por formulario')]/..",
		"//span[contains(text(), 'Búsqueda de licitaciones por formulario')]",
	}
	
			for _, selector := range selectors {
		log.Printf("  Trying selector: %s", selector)
		licitacionesLink, err = s.driver.FindElement(selenium.ByXPATH, selector)
		if err == nil {
			// Get the tag name to understand what type of element we found
			tagName, err := licitacionesLink.TagName()
			if err == nil {
				log.Printf("✅ Found element with tag: <%s>", tagName)
			}
			
			// Verify this is the right link by checking its text or href
			text, err := licitacionesLink.Text()
			if err == nil {
				log.Printf("✅ Found potential link with text: '%s'", text)
				if strings.Contains(strings.ToLower(text), "licitaciones") || 
				   strings.Contains(strings.ToLower(text), "búsqueda") ||
				   strings.Contains(strings.ToLower(text), "formulario") {
					log.Printf("✅ Confirmed Licitaciones link: %s", text)
					return licitacionesLink, nil
				}
			}
			
			// Also check href attribute
			href, err := licitacionesLink.GetAttribute("href")
			if err == nil {
				log.Printf("✅ Found potential link with href: '%s'", href)
				if strings.Contains(strings.ToLower(href), "formulario") || 
				   strings.Contains(strings.ToLower(href), "busqueda") ||
				   strings.Contains(strings.ToLower(href), "licitaciones") {
					log.Printf("✅ Confirmed Licitaciones link by href: %s", href)
					return licitacionesLink, nil
				}
			}
			
			// If we found a span, try to find its parent link
			if tagName == "span" {
				log.Println("Found span element, looking for parent link...")
				parentLink, err := s.driver.FindElement(selenium.ByXPATH, "//span[contains(text(), 'Búsqueda de licitaciones por formulario')]/parent::a")
				if err == nil {
					log.Println("✅ Found parent link for span")
					return parentLink, nil
				}
			}
		}
	}
	
	// Strategy 3: Try to find any clickable element that might lead to the search form
	log.Println("Strategy 3: Looking for any clickable elements...")
	allLinks, err := s.driver.FindElements(selenium.ByTagName, "a")
	if err == nil {
		log.Printf("Found %d links on the page", len(allLinks))
		for i, link := range allLinks {
			text, err := link.Text()
			if err == nil {
				text = strings.TrimSpace(text)
				if text != "" {
					log.Printf("  Link %d: '%s'", i, text)
					if strings.Contains(strings.ToLower(text), "licitaciones") || 
					   strings.Contains(strings.ToLower(text), "búsqueda") ||
					   strings.Contains(strings.ToLower(text), "formulario") {
						log.Printf("✅ Found Licitaciones link by text: %s", text)
						return link, nil
					}
				}
			}
		}
	}
	
	// Strategy 4: Try to get page source and analyze it
	log.Println("Strategy 4: Analyzing page source...")
	pageSource, err := s.driver.PageSource()
	if err == nil {
		log.Printf("Page source length: %d characters", len(pageSource))
		// Look for the specific ID in the page source
		if strings.Contains(pageSource, "viewns_Z7_AVEQAI930OBRD02JPMTPG21004_:form1:linkFormularioBusqueda") {
			log.Println("✅ Found the ID in page source, trying again...")
			licitacionesLink, err = s.driver.FindElement(selenium.ByID, "viewns_Z7_AVEQAI930OBRD02JPMTPG21004_:form1:linkFormularioBusqueda")
			if err == nil {
				return licitacionesLink, nil
			}
		}
		
		// Look for any link containing "licitaciones" or "formulario"
		if strings.Contains(strings.ToLower(pageSource), "licitaciones") {
			log.Println("✅ Found 'licitaciones' in page source")
		}
		if strings.Contains(strings.ToLower(pageSource), "formulario") {
			log.Println("✅ Found 'formulario' in page source")
		}
	}
	
	return nil, fmt.Errorf("could not find Licitaciones link with any strategy")
}




// GetSessionID returns the current session ID
func (s *SeleniumScraper) GetSessionID() string {
	return s.sessionID
}

// TakeScreenshotWithDescription takes a screenshot with a custom description
func (s *SeleniumScraper) TakeScreenshotWithDescription(description string) error {
	// Create a clean filename from the description
	cleanDescription := strings.ReplaceAll(description, " ", "_")
	cleanDescription = strings.ReplaceAll(cleanDescription, "-", "_")
	cleanDescription = strings.ReplaceAll(cleanDescription, ".", "_")
	cleanDescription = strings.ReplaceAll(cleanDescription, ":", "_")
	
	filename := fmt.Sprintf("%s.png", cleanDescription)
	return s.TakeScreenshot(filename)
}

// TakeScreenshot takes a screenshot for debugging
func (s *SeleniumScraper) TakeScreenshot(filename string) error {
	bytes, err := s.driver.Screenshot()
	if err != nil {
		return fmt.Errorf("failed to take screenshot: %w", err)
	}

	// Create screenshots directory if it doesn't exist
	screenshotsDir := fmt.Sprintf("screenshots/%s", s.sessionID)
	if err := os.MkdirAll(screenshotsDir, 0755); err != nil {
		return fmt.Errorf("failed to create screenshots directory: %w", err)
	}

	// Generate timestamp for unique naming (human-readable format)
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

	log.Printf("📸 Screenshot saved to: %s", fullPath)
	return nil
}

// DebugPageStructure analyzes and logs the page structure for debugging
func (s *SeleniumScraper) DebugPageStructure() error {
	log.Println("=== DEBUGGING PAGE STRUCTURE ===")
	
	// Get current URL
	currentURL, err := s.driver.CurrentURL()
	if err == nil {
		log.Printf("Current URL: %s", currentURL)
	}
	
	// Get page title
	title, err := s.driver.Title()
	if err == nil {
		log.Printf("Page title: %s", title)
	}
	
	// Find all links on the page
	links, err := s.driver.FindElements(selenium.ByTagName, "a")
	if err == nil {
		log.Printf("Found %d links on the page", len(links))
		for i, link := range links {
			if i >= 20 { // Limit to first 20 links
				log.Printf("... and %d more links", len(links)-20)
				break
			}
			
			text, err := link.Text()
			if err == nil {
				text = strings.TrimSpace(text)
				if text != "" {
					href, _ := link.GetAttribute("href")
					log.Printf("  Link %d: '%s' -> %s", i, text, href)
				}
			}
		}
	}
	
	// Find all buttons on the page
	buttons, err := s.driver.FindElements(selenium.ByTagName, "button")
	if err == nil {
		log.Printf("Found %d buttons on the page", len(buttons))
		for i, button := range buttons {
			if i >= 10 { // Limit to first 10 buttons
				log.Printf("... and %d more buttons", len(buttons)-10)
				break
			}
			
			text, err := button.Text()
			if err == nil {
				text = strings.TrimSpace(text)
				if text != "" {
					log.Printf("  Button %d: '%s'", i, text)
				}
			}
		}
	}
	
	// Find all input elements
	inputs, err := s.driver.FindElements(selenium.ByTagName, "input")
	if err == nil {
		log.Printf("Found %d input elements on the page", len(inputs))
		for i, input := range inputs {
			if i >= 10 { // Limit to first 10 inputs
				log.Printf("... and %d more inputs", len(inputs)-10)
				break
			}
			
			inputType, _ := input.GetAttribute("type")
			placeholder, _ := input.GetAttribute("placeholder")
			name, _ := input.GetAttribute("name")
			id, _ := input.GetAttribute("id")
			log.Printf("  Input %d: type=%s, name=%s, id=%s, placeholder=%s", i, inputType, name, id, placeholder)
		}
	}
	
	// Look for specific elements we're interested in
	log.Println("=== LOOKING FOR SPECIFIC ELEMENTS ===")
	
	// Try to find the specific ID
	_, err = s.driver.FindElement(selenium.ByID, "viewns_Z7_AVEQAI930OBRD02JPMTPG21004_:form1:linkFormularioBusqueda")
	if err == nil {
		log.Println("✅ Found the specific ID: viewns_Z7_AVEQAI930OBRD02JPMTPG21004_:form1:linkFormularioBusqueda")
	} else {
		log.Printf("❌ Could not find the specific ID: %v", err)
	}
	
	// Look for any element containing "licitaciones"
	licitacionesElements, err := s.driver.FindElements(selenium.ByXPATH, "//*[contains(text(), 'Licitaciones')]")
	if err == nil {
		log.Printf("Found %d elements containing 'Licitaciones'", len(licitacionesElements))
		for i, elem := range licitacionesElements {
			if i >= 5 { // Limit to first 5
				log.Printf("... and %d more", len(licitacionesElements)-5)
				break
			}
			text, _ := elem.Text()
			tagName, _ := elem.TagName()
			log.Printf("  Element %d: <%s> '%s'", i, tagName, strings.TrimSpace(text))
		}
	}
	
	// Look for any element containing "formulario"
	formularioElements, err := s.driver.FindElements(selenium.ByXPATH, "//*[contains(text(), 'formulario')]")
	if err == nil {
		log.Printf("Found %d elements containing 'formulario'", len(formularioElements))
		for i, elem := range formularioElements {
			if i >= 5 { // Limit to first 5
				log.Printf("... and %d more", len(formularioElements)-5)
				break
			}
			text, _ := elem.Text()
			tagName, _ := elem.TagName()
			log.Printf("  Element %d: <%s> '%s'", i, tagName, strings.TrimSpace(text))
		}
	}
	
	log.Println("=== END DEBUGGING ===")
	return nil
} 

// GetScreenshotsDirectory returns the path to the current session's screenshots directory
func (s *SeleniumScraper) GetScreenshotsDirectory() string {
	return fmt.Sprintf("screenshots/%s", s.sessionID)
}

// ListScreenshots returns a list of all screenshots taken in this session
func (s *SeleniumScraper) ListScreenshots() ([]string, error) {
	screenshotsDir := s.GetScreenshotsDirectory()
	
	// Check if directory exists
	if _, err := os.Stat(screenshotsDir); os.IsNotExist(err) {
		return []string{}, nil // Return empty list if directory doesn't exist
	}
	
	// Read directory contents
	files, err := os.ReadDir(screenshotsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read screenshots directory: %w", err)
	}
	
	var screenshots []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".png") {
			screenshots = append(screenshots, file.Name())
		}
	}
	
	// Sort screenshots by name (which includes timestamp, so they'll be chronological)
	sort.Strings(screenshots)
	
	return screenshots, nil
}

// GetSessionInfo returns information about the current scraping session
func (s *SeleniumScraper) GetSessionInfo() map[string]interface{} {
	screenshots, _ := s.ListScreenshots()
	
	return map[string]interface{}{
		"session_id":           s.sessionID,
		"screenshots_directory": s.GetScreenshotsDirectory(),
		"screenshots_count":    len(screenshots),
		"screenshots_list":     screenshots,
		"session_started":      s.sessionID[8:], 
	}
} 