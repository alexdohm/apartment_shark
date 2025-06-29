# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Apartment Hunter is a Go application that continuously scrapes German housing websites for new apartment listings and sends notifications via Telegram. It targets specific Berlin neighborhoods based on configurable zip codes and price filters.

## Development Commands

### Build and Run
```bash
# Build the application
go build -o app .

# Run locally
go run cmd/main.go

# Build for Linux (production)
GOOS=linux GOARCH=amd64 go build -o app .
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests for specific package
go test ./internal/telegram
go test ./internal/store

# Run tests with verbose output
go test -v ./...
```

### Dependencies
```bash
# Download dependencies
go mod download

# Update dependencies
go mod tidy

# View module information
go mod list -m all
```

### Deployment
```bash
# Deploy to production (requires Docker and Nomad setup)
./deploy.sh
```

## Architecture

### Core Components

**Main Application Flow:**
- `cmd/main.go` - Entry point that initializes scrapers and runs them concurrently
- Each scraper runs in its own goroutine with randomized timing to avoid detection

**Scraper Architecture:**
- `internal/scraping/common/scraper.go` - Base scraper interface and implementation
- `internal/scraping/factory/factory.go` - Factory pattern for creating company-specific scrapers
- `internal/scraping/companies/*/` - Individual company scraper implementations

**Key Services:**
- `internal/telegram/` - Telegram bot integration for notifications
- `internal/http/client.go` - HTTP client with timeout configuration
- `internal/store/store.go` - In-memory state management to track processed listings
- `internal/bot/` - Header generation and delay utilities for web scraping
- `internal/config/config.go` - Configuration constants (contains hardcoded Telegram credentials)

### Scraper System

**Architecture (Decoupled Design):**
The application uses a clean separation pattern:
1. **Scrapers** - Return standardized `Listing` structs without handling notifications
2. **Main Orchestration** - Coordinates scrapers, deduplication, and notifications using existing `telegram.Notifier`
3. **State Management** - Accessed through scraper interface to prevent duplicate notifications

**Implementation:**
- `DefaultScraperFactory` creates scrapers for different companies
- Each company scraper implements the `common.Scraper` interface
- `BaseScraper` provides shared functionality (HTTP client, headers, state)
- Company-specific scraping logic is injected as `ScrapingFunc`

**Current Interface:**
```go
// Scraper interface with state access for deduplication
type Scraper interface {
    GetName() string
    Scrape(ctx context.Context) ([]Listing, error)
    GetState() *store.ScraperState
}

// Standardized listing structure
type Listing struct {
    ID      string
    Company string
    Price   string
    Size    string
    Address string
    URL     string
}

// Helper method for telegram conversion
func (l Listing) ToTelegramInfo() *telegram.TelegramInfo
```

### Configuration

All configuration is hardcoded in `internal/config/config.go`:
- Telegram bot credentials (BaseURL, BotToken, ChatID)
- Target URLs for each housing company
- Search filters (zip codes, price ranges, minimum square meters)

### State Management

`internal/store/ScraperState` maintains in-memory tracking of processed listings to prevent duplicate notifications. State is not persisted between application restarts.

## Key Files for Modifications

- `cmd/main.go:17-23` - Enable/disable scrapers by modifying `scrapersTypes` array
- `internal/config/config.go` - Update search filters, URLs, or Telegram configuration
- `internal/scraping/companies/*/scraper.go` - Company-specific scraping logic
- `internal/scraping/common/listing.go` - Standardized listing structure and telegram conversion

## Current Implementation Details

**Decoupled Design Implementation:**
The codebase now implements the clean separation pattern:

**Data Flow:**
```
Scrapers → []common.Listing → State Check → telegram.Notifier → Telegram API
```

**Key Implementation Points:**
1. **Scrapers** return `[]common.Listing` structs (implemented in `internal/scraping/companies/*/scraper.go`)
2. **State Management** accessed via `scraper.GetState()` in `cmd/main.go:72`
3. **Notifications** handled by passing `telegram.Notifier` to `startScraper()` function
4. **Conversion** from `common.Listing` to `telegram.TelegramInfo` via `listing.ToTelegramInfo()`

**Benefits Achieved:**
- **Clean Separation**: Scrapers focus solely on data extraction
- **Testable**: Can mock `telegram.Notifier` interface for unit tests
- **Extensible**: Easy to add new notification channels without changing scrapers
- **Maintainable**: Notification logic separated from scraping logic

**Testing:**
```go
// Example unit test structure
func TestScraper(t *testing.T) {
    mockNotifier := &MockTelegramNotifier{}
    // Test scraper logic without network calls
}
```

## Production Deployment

The application is containerized and deployed using:
- Docker multi-stage build (Dockerfile)
- Nomad orchestration (apartment-hunter.nomad)
- Automated deployment via deploy.sh script