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
# Run all tests (unit tests only)
go test ./...

# Run tests for specific package
go test ./internal/telegram
go test ./internal/store
go test ./internal/scraping/common

# Run tests with verbose output
go test -v ./...

# Run integration tests (real endpoint testing)
go test -tags=integration ./internal/scraping/companies/dewego
go test -tags=integration ./internal/scraping/companies/howoge
go test -tags=integration ./internal/scraping/companies/gewobag
go test -tags=integration ./internal/scraping/companies/stadtundland
go test -tags=integration ./internal/scraping/companies/wbm

# Run integration tests with timeout (recommended)
go test -tags=integration -timeout=60s ./internal/scraping/companies/...

# Skip integration tests (default behavior)
go test -short ./...
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

**Architecture (Function-Based Design):**
The application uses a clean separation pattern:
1. **Scrapers** - Return standardized `Listing` structs via standalone `FetchListings` functions
2. **Main Orchestration** - Coordinates scrapers, deduplication, and notifications using `telegram.Client`
3. **State Management** - Accessed through scraper interface to prevent duplicate notifications

**Implementation:**
- `DefaultScraperFactory` creates `BaseScraper` instances for different companies
- Each company implements a `FetchListings(ctx, *BaseScraper) ([]Listing, error)` function
- `BaseScraper` provides shared functionality (HTTP client, headers, state) and handles error formatting
- Company-specific scraping logic is contained in standalone functions

**Scraper Types by Technology:**
- **HTML Scrapers**: Dewego, Gewobag, WBM (use goquery for CSS selector-based parsing)
- **JSON API Scrapers**: Howoge (form POST → JSON), Stadt Und Land (JSON POST → JSON)

### Configuration

All configuration is hardcoded in `internal/config/config.go`:
- Telegram bot credentials (BaseURL, BotToken, ChatID)
- Target URLs for each housing company
- Search filters (zip codes, price ranges, minimum square meters)

### State Management

`internal/store/ScraperState` maintains in-memory tracking of processed listings to prevent duplicate notifications. State is not persisted between application restarts.

### Testing Architecture

**Unit Tests** - Test business logic with mocks:
- `internal/scraping/common/*_test.go` - Core scraping logic
- `internal/telegram/*_test.go` - Telegram client functionality
- Use HTTP mocks from `internal/http/mock/`

**Integration Tests** - Test against real endpoints (with build tags):
- `internal/scraping/companies/*/scraper_test.go` - Real endpoint validation
- Run with `go test -tags=integration`
- Test endpoint reachability, HTML/JSON structure, and CSS selectors/API fields

## Key Files for Modifications

- `cmd/main.go:17-23` - Enable/disable scrapers by modifying `scrapersTypes` array
- `internal/config/config.go` - Update search filters, URLs, or Telegram configuration
- `internal/scraping/companies/*/scraper.go` - Company-specific `FetchListings` functions
- `internal/scraping/common/listing.go` - Standardized listing structure and telegram conversion
- `internal/telegram/client.go` - Telegram client (replaces old notifier/send separation)

## Implementation Patterns

**Scraper Function Structure:**
All company scrapers follow the same pattern in their `FetchListings` function:
1. Build request parameters (form data or JSON)
2. Make HTTP request via `base.HTTPClient`
3. Parse response (HTML with goquery or JSON unmarshaling)
4. Extract and normalize listing data
5. Return `[]common.Listing` structs

**Data Flow:**
```
FetchListings → []common.Listing → BaseScraper.Scrape (error formatting) → State Check → telegram.Client → Telegram API
```

**Error Handling:**
- `BaseScraper.Scrape()` wraps company function errors with standardized format
- Integration tests validate endpoint changes don't break scrapers
- State management prevents duplicate notifications

**Pagination Support:**
Some scrapers (like Dewego) implement pagination to retrieve all available listings across multiple pages with controlled delays between requests.

## Production Deployment

The application is containerized and deployed using:
- Docker multi-stage build (Dockerfile)
- Nomad orchestration (apartment-hunter.nomad)
- Automated deployment via deploy.sh script