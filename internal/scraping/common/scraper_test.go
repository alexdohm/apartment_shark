package common

import (
	"apartmenthunter/internal/http"
	"apartmenthunter/internal/http/mock"
	"apartmenthunter/internal/store"
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func createTestBaseScraper(httpClient http.HTTPClient, name string, scrapingFunc ScrapingFunc) *BaseScraper {
	state := store.NewScraperState()
	return NewBaseScraper(httpClient, state, name, scrapingFunc)
}

func createSuccessfulScrapingFunc(listings []Listing) ScrapingFunc {
	return func(ctx context.Context, base *BaseScraper) ([]Listing, error) {
		return listings, nil
	}
}

func createFailingScrapingFunc(err error) ScrapingFunc {
	return func(ctx context.Context, base *BaseScraper) ([]Listing, error) {
		return nil, err
	}
}

// TestNewBaseScraper tests the BaseScraper constructor
func TestBaseScraper_GetName(t *testing.T) {
	tests := []struct {
		name        string
		scraperName string
	}{
		{"normal name", "TestCompany"},
		{"empty name", ""},
		{"name with spaces", "Test Company Name"},
		{"name with special chars", "Test-Company_123"},
		{"unicode name", "Tëst Çömpäny"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scraper := createTestBaseScraper(mock.NewHTTPClient(), tt.scraperName, nil)

			result := scraper.GetName()
			if result != tt.scraperName {
				t.Errorf("GetName() = %v, want %v", result, tt.scraperName)
			}
		})
	}
}

// TestBaseScraper_GetState tests the GetState method
func TestBaseScraper_GetState(t *testing.T) {
	tests := []struct {
		name  string
		state *store.ScraperState
	}{
		{"valid state", store.NewScraperState()},
		{"nil state", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scraper := NewBaseScraper(mock.NewHTTPClient(), tt.state, "TestScraper", nil)

			result := scraper.GetState()
			if result != tt.state {
				t.Errorf("GetState() = %v, want %v", result, tt.state)
			}
		})
	}
}

// TestBaseScraper_Scrape tests the main Scrape method
func TestBaseScraper_Scrape(t *testing.T) {
	tests := []struct {
		name            string
		scraperName     string
		scrapingFunc    ScrapingFunc
		wantErr         bool
		wantCount       int
		wantErrContains string
	}{
		{
			name:        "successful scraping with results",
			scraperName: "TestCompany",
			scrapingFunc: createSuccessfulScrapingFunc([]Listing{
				{ID: "1", Company: "Test", Price: "800", Size: "45", Address: "Test St", URL: "http://test.com"},
				{ID: "2", Company: "Test", Price: "900", Size: "50", Address: "Test Ave", URL: "http://test2.com"},
			}),
			wantErr:   false,
			wantCount: 2,
		},
		{
			name:         "successful scraping with empty results",
			scraperName:  "TestCompany",
			scrapingFunc: createSuccessfulScrapingFunc([]Listing{}),
			wantErr:      false,
			wantCount:    0,
		},
		{
			name:            "scraping function returns error",
			scraperName:     "TestCompany",
			scrapingFunc:    createFailingScrapingFunc(errors.New("network timeout")),
			wantErr:         true,
			wantErrContains: "fetching TestCompany listings: network timeout",
		},
		{
			name:            "scraping function returns wrapped error",
			scraperName:     "ErrorCompany",
			scrapingFunc:    createFailingScrapingFunc(errors.New("HTTP 500: server error")),
			wantErr:         true,
			wantErrContains: "fetching ErrorCompany listings: HTTP 500: server error",
		},
		{
			name:        "scraping function returns nil listings with nil error",
			scraperName: "TestCompany",
			scrapingFunc: func(ctx context.Context, base *BaseScraper) ([]Listing, error) {
				return nil, nil
			},
			wantErr:   false,
			wantCount: 0,
		},
		{
			name:         "scraping function with large result set",
			scraperName:  "TestCompany",
			scrapingFunc: createSuccessfulScrapingFunc(make([]Listing, 1000)), // 1000 empty listings
			wantErr:      false,
			wantCount:    1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scraper := createTestBaseScraper(mock.NewHTTPClient(), tt.scraperName, tt.scrapingFunc)

			ctx := context.Background()
			result, err := scraper.Scrape(ctx)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Scrape() error = nil, want error")
				} else if tt.wantErrContains != "" && !strings.Contains(err.Error(), tt.wantErrContains) {
					t.Errorf("Scrape() error = %v, want to contain %v", err.Error(), tt.wantErrContains)
				}
				return
			}

			if err != nil {
				t.Errorf("Scrape() unexpected error: %v", err)
				return
			}

			if len(result) != tt.wantCount {
				t.Errorf("Scrape() result count = %v, want %v", len(result), tt.wantCount)
			}
		})
	}
}

// TestBaseScraper_Scrape_ContextHandling tests context cancellation and timeout
func TestBaseScraper_Scrape_ContextHandling(t *testing.T) {
	tests := []struct {
		name            string
		setupContext    func() (context.Context, context.CancelFunc)
		scrapingFunc    ScrapingFunc
		wantErr         bool
		wantErrContains string
	}{
		{
			name: "context already cancelled",
			setupContext: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // Cancel immediately
				return ctx, cancel
			},
			scrapingFunc: func(ctx context.Context, base *BaseScraper) ([]Listing, error) {
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				default:
					return []Listing{}, nil
				}
			},
			wantErr:         true,
			wantErrContains: "context canceled",
		},
		{
			name: "context timeout",
			setupContext: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 10*time.Millisecond)
			},
			scrapingFunc: func(ctx context.Context, base *BaseScraper) ([]Listing, error) {
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(50 * time.Millisecond): // Longer than timeout
					return []Listing{}, nil
				}
			},
			wantErr:         true,
			wantErrContains: "deadline exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scraper := createTestBaseScraper(mock.NewHTTPClient(), "TestCompany", tt.scrapingFunc)

			ctx, cancel := tt.setupContext()
			defer cancel()

			result, err := scraper.Scrape(ctx)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Scrape() error = nil, want error")
				} else if tt.wantErrContains != "" && !strings.Contains(err.Error(), tt.wantErrContains) {
					t.Errorf("Scrape() error = %v, want to contain %v", err.Error(), tt.wantErrContains)
				}
			} else {
				if err != nil {
					t.Errorf("Scrape() unexpected error: %v", err)
				}
				if tt.name == "context with value" && len(result) != 1 {
					t.Errorf("Scrape() result count = %v, want 1", len(result))
				}
			}
		})
	}
}

// TestBaseScraper_Scrape_NilScrapingFunc tests edge case with nil scraping function
func TestBaseScraper_Scrape_NilScrapingFunc(t *testing.T) {
	scraper := createTestBaseScraper(mock.NewHTTPClient(), "TestCompany", nil)
	ctx := context.Background()

	// This should panic, so we test with recover
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Scrape() with nil scrapingFunc should panic")
		}
	}()

	_, _ = scraper.Scrape(ctx)
}
