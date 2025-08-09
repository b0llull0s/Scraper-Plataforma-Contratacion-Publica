# LED Screen Contract Scraper

A Go-based scraper tool to monitor Spanish public procurement contracts for LED screens from [contrataciondelestado.es](https://contrataciondelestado.es).

## Features

- **Scraping workflow** for CPV "32351200" (navigate → fill CPV → add → search → wait → extract)
- **Two run modes**: visible browser (`--scrape-selenium`) and headless (`--scrape-cli`)
- **Status change tracking** with `status_changes` history and recent changes API/UI
- **SQLite** persistence and simple CRUD (delete all / delete one)
- **Email notifications** for new contracts
- **Web dashboard** to view/search contracts and see recent status changes
- **Screenshots** per session for debugging (saved under `screenshots/<session_id>`)

## Project Structure

```
scraper/
├── cmd/
│   └── main.go              # CLI entry point
├── internal/
│   ├── scraper/             # Unified core + Selenium drivers (visible & headless)
│   ├── storage/             # SQLite schema & queries (contracts + status_changes)
│   ├── notification/        # Email alerts
│   └── dashboard/           # Web interface (inline templates)
├── go.mod                   # Go module file
└── README.md                # This file
```

## Quick Start

### Prerequisites

- Go 1.19 or later
- SQLite3
- For Selenium modes: Selenium/ChromeDriver listening on port 4444, 4445, or 4446

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd Scraper-Plataforma-Contratacion
```

2. Build the project:
```bash
go build -o scraper cmd/main.go
```

### Configuration

Set up environment variables for email notifications:

```bash
export SMTP_HOST="smtp.gmail.com"
export SMTP_PORT="587"
export SMTP_USERNAME="your-email@gmail.com"
export SMTP_PASSWORD="your-app-password"
export FROM_EMAIL="your-email@gmail.com"
export TO_EMAIL="recipient@example.com"
```

### Usage

#### Test Connection
Test if the scraper can reach and operate on the target website (requires Selenium running):
```bash
./scraper --test
```

#### Test Email Configuration
```bash
./scraper --test-email
```

#### Run Scraper
Scrape for LED screen contracts:

Visible browser (Selenium):
```bash
./scraper --scrape-selenium --db contracts.db
```

Headless (CLI mode, Selenium in headless Chrome):
```bash
./scraper --scrape-cli --db contracts.db
```

Optional Selenium debug (navigates and inspects page; saves screenshots):
```bash
./scraper --debug-selenium
```

#### Start Dashboard
```bash
./scraper --serve --port 8080 --db contracts.db
```
Open http://localhost:8080

#### Other Options
```bash
./scraper --db contracts.db    # Database file path (default: contracts.db)
./scraper --port 3000          # Dashboard port (default: 8080)
```

## Dashboard Features

- Real-time contract list with search
- Statistics (total) and recent status changes panel
- Contract details: ID, description, amount, status, submission date, contracting body, scraped time
- Document links (Pliego/Anuncio) when available
- Delete all contracts / delete a single contract
- Status change history page at `/history`

## Building for Different Platforms

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o scraper-linux cmd/main.go

# Windows
GOOS=windows GOARCH=amd64 go build -o scraper.exe cmd/main.go

# macOS
GOOS=darwin GOARCH=amd64 go build -o scraper-mac cmd/main.go
```
