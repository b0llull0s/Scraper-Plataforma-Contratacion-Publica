package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"scraper/internal/dashboard"
	"scraper/internal/notification"
	"scraper/internal/scraper"
	"scraper/internal/storage"
)

func main() {
	// Define command line flags
	var (
		testConnection = flag.Bool("test", false, "Test connection to the website")
		testEmail      = flag.Bool("test-email", false, "Test email configuration")
		scrapeSelenium = flag.Bool("scrape-selenium", false, "Run the Selenium-based scraper (requires Selenium server)")
		scrapeCLI      = flag.Bool("scrape-cli", false, "Run the CLI-only scraper (headless Selenium, requires Selenium server)")
		debugSelenium  = flag.Bool("debug-selenium", false, "Debug Selenium page structure (navigates to page and analyzes it)")
		serve          = flag.Bool("serve", false, "Start the web dashboard")
		dbPath         = flag.String("db", "contracts.db", "Database file path")
		port           = flag.String("port", "8080", "Dashboard port")
	)
	flag.Parse()

	// Initialize storage
	store, err := storage.NewStorage(*dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()

	// Initialize notifier (you'll need to set these environment variables)
	notifier := notification.NewNotifier(
		os.Getenv("SMTP_HOST"),
		os.Getenv("SMTP_PORT"),
		os.Getenv("SMTP_USERNAME"),
		os.Getenv("SMTP_PASSWORD"),
		os.Getenv("FROM_EMAIL"),
		[]string{os.Getenv("TO_EMAIL")}, // You can add multiple emails separated by comma
	)

	// Handle different commands
	switch {
	case *testConnection:
		// Test connection using CLI scraper (headless mode)
		cliScraper, err := scraper.NewScraper(scraper.ScraperTypeCLI)
		if err != nil {
			log.Fatalf("Failed to create CLI scraper for connection test: %v", err)
		}
		defer cliScraper.Close()
		
		// Test by trying to navigate to the base URL
		if err := cliScraper.NavigateToSearchForm(); err != nil {
			log.Fatalf("Connection test failed: %v", err)
		}
		fmt.Println("✅ Connection test successful!")

	case *testEmail:
		if err := notifier.TestConnection(); err != nil {
			log.Fatalf("Email test failed: %v", err)
		}
		fmt.Println("✅ Email configuration test successful!")

	case *scrapeSelenium:
		fmt.Println("🔍 Starting unified scraper (Selenium mode)...")
		
		// Use the unified scraping function with Selenium mode
		contracts, err := scraper.ScrapeContracts(scraper.ScraperTypeSelenium)
		if err != nil {
			log.Fatalf("Selenium scraping failed: %v", err)
		}

		fmt.Printf("📊 Found %d contracts with Selenium\n", len(contracts))
		processContracts(contracts, store, notifier)

	case *scrapeCLI:
		fmt.Println("🔍 Starting unified scraper (CLI mode)...")
		
		// Create CLI scraper instance
		cliScraper, err := scraper.NewScraper(scraper.ScraperTypeCLI)
		if err != nil {
			log.Fatalf("Failed to create CLI scraper: %v", err)
		}
		defer cliScraper.Close()

		// Use the unified scraping workflow
		contracts, err := scraper.ScrapeContractsWithScraper(cliScraper)
		if err != nil {
			log.Fatalf("CLI scraping failed: %v", err)
		}

		// Extract ALL contracts for status change detection
		allContracts, err := cliScraper.ExtractAllContracts()
		if err != nil {
			log.Printf("Warning: Failed to extract all contracts for status checking: %v", err)
			allContracts = []scraper.Contract{} // Empty slice if failed
		}

		// Enhance contracts with document links (Pliego and Anuncio)
		fmt.Println("📄 Enhancing contracts with document links...")
		coreScraper := scraper.NewCoreScraper()
		enhancedContracts, err := coreScraper.EnhanceContractsWithDocumentLinks(contracts, cliScraper, store)
		if err != nil {
			log.Printf("Warning: Failed to enhance contracts with document links: %v", err)
			enhancedContracts = contracts // Use original contracts if enhancement fails
		}

		fmt.Printf("📊 Found %d contracts with CLI scraper\n", len(enhancedContracts))
		fmt.Printf("📋 Found %d total contracts for status change detection\n", len(allContracts))
		processContractsWithStatusCheck(enhancedContracts, allContracts, store, notifier)

	case *debugSelenium:
		fmt.Println("🔍 Starting Selenium debug mode...")
		
		// Initialize Selenium scraper for debugging
		seleniumScraper, err := scraper.NewSeleniumScraper()
		if err != nil {
			log.Fatalf("Failed to initialize Selenium scraper: %v", err)
		}
		defer seleniumScraper.Close()

		// Navigate to the main page
		log.Println("Navigating to main licitaciones page...")
		if err := seleniumScraper.GetDriver().Get(seleniumScraper.GetBaseURL() + "/wps/portal/licitaciones"); err != nil {
			log.Fatalf("Failed to navigate to licitaciones page: %v", err)
		}

		log.Println("✅ Successfully navigated to licitaciones page")
		log.Println("⏳ Waiting 10 seconds for page to fully load...")
		time.Sleep(10 * time.Second)

		// Take a screenshot
		if err := seleniumScraper.TakeScreenshot("debug_page.png"); err != nil {
			log.Printf("Warning: Failed to take screenshot: %v", err)
		}

		// Debug the page structure
		log.Println("🔍 Debugging page structure...")
		if err := seleniumScraper.DebugPageStructure(); err != nil {
			log.Printf("Warning: Page structure debugging failed: %v", err)
		}

		// Try to find and click the Licitaciones link
		log.Println("🔍 Looking for Licitaciones link...")
		licitacionesLink, err := seleniumScraper.FindLicitacionesLink()
		if err != nil {
			log.Printf("❌ Could not find Licitaciones link: %v", err)
		} else {
			log.Println("✅ Found Licitaciones link, clicking...")
			if err := licitacionesLink.Click(); err != nil {
				log.Printf("❌ Failed to click Licitaciones link: %v", err)
			} else {
				log.Println("✅ Successfully clicked Licitaciones link")
				log.Println("⏳ Waiting 10 seconds for search form to load...")
				time.Sleep(10 * time.Second)

				// Take a screenshot of the search form
				if err := seleniumScraper.TakeScreenshot("debug_search_form.png"); err != nil {
					log.Printf("Warning: Failed to take screenshot: %v", err)
				}

				// Debug the search form page structure
				log.Println("🔍 Debugging search form page structure...")
				if err := seleniumScraper.DebugPageStructure(); err != nil {
					log.Printf("Warning: Search form page structure debugging failed: %v", err)
				}
			}
		}

		fmt.Println("✅ Debug mode completed. Check the logs and screenshots for details.")

	case *serve:
		fmt.Printf("🌐 Starting dashboard on port %s...\n", *port)
		dashboard := dashboard.NewDashboard(store, *port)
		if err := dashboard.Start(); err != nil {
			log.Fatalf("Failed to start dashboard: %v", err)
		}

	default:
		fmt.Println("LED Screen Contract Scraper")
		fmt.Println("Usage:")
		fmt.Println("  --test            Test connection to the website")
		fmt.Println("  --test-email      Test email configuration")
		fmt.Println("  --scrape-selenium Run the Selenium-based scraper (requires Selenium server)")
		fmt.Println("  --scrape-cli      Run the CLI-only scraper (headless Selenium, requires Selenium server)")
		fmt.Println("  --debug-selenium  Debug Selenium page structure (navigates to page and analyzes it)")
		fmt.Println("  --serve           Start the web dashboard")
		fmt.Println("  --db PATH         Database file path (default: contracts.db)")
		fmt.Println("  --port PORT       Dashboard port (default: 8080)")
		fmt.Println()
		fmt.Println("Environment variables needed for email:")
		fmt.Println("  SMTP_HOST, SMTP_PORT, SMTP_USERNAME, SMTP_PASSWORD")
		fmt.Println("  FROM_EMAIL, TO_EMAIL")
		fmt.Println()
		fmt.Println("For Selenium scraper, you need to:")
		fmt.Println("  1. Install Selenium server: docker run -d -p 4444:4444 selenium/standalone-chrome")
		fmt.Println("  2. Or install ChromeDriver and run: chromedriver --port=4444")
	}
}

// processContracts handles the common logic for processing scraped contracts
func processContracts(contracts []scraper.Contract, store *storage.Storage, notifier *notification.Notifier) {
	if len(contracts) > 0 {
		// Get new contracts
		newContracts, err := store.GetNewContracts(contracts)
		if err != nil {
			log.Fatalf("Failed to check for new contracts: %v", err)
		}

		fmt.Printf("🆕 Found %d new contracts\n", len(newContracts))

		// Save all contracts (this will also detect status changes)
		if err := store.SaveContracts(contracts); err != nil {
			log.Fatalf("Failed to save contracts: %v", err)
		}

		// Send notification for new contracts
		if len(newContracts) > 0 {
			if err := notifier.SendNewContractsNotification(newContracts); err != nil {
				log.Printf("Warning: Failed to send notification: %v", err)
			} else {
				fmt.Println("📧 Notification sent for new contracts")
			}
		}
	}

	// Show total count
	count, err := store.GetContractCount()
	if err != nil {
		log.Printf("Warning: Failed to get contract count: %v", err)
	} else {
		fmt.Printf("💾 Total contracts in database: %d\n", count)
	}
}

// processContractsWithStatusCheck handles contracts and status changes
func processContractsWithStatusCheck(contracts []scraper.Contract, allContracts []scraper.Contract, store *storage.Storage, notifier *notification.Notifier) {
	// First, check for status changes in existing contracts
	if len(allContracts) > 0 {
		if err := store.CheckAndUpdateStatusChanges(allContracts); err != nil {
			log.Printf("Warning: Failed to check status changes: %v", err)
		}
	}

	// Then process new contracts
	processContracts(contracts, store, notifier)

	// Check for status changes
	statusChanges, err := store.GetRecentStatusChanges()
	if err != nil {
		log.Printf("Warning: Failed to get status changes: %v", err)
	} else if len(statusChanges) > 0 {
		fmt.Printf("🔄 Found %d status changes:\n", len(statusChanges))
		for _, change := range statusChanges {
			fmt.Printf("   • %s: %s → %s (%s)\n", change.ContractID, change.OldStatus, change.NewStatus, change.ChangedAt)
		}
	}
} 