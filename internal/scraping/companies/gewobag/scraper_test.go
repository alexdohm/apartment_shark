package gewobag

import (
	"apartmenthunter/internal/config"
	"apartmenthunter/internal/http"
	"apartmenthunter/internal/scraping/common"
	"apartmenthunter/internal/store"
	"context"
	"strings"
	"testing"
	"time"
)

// TestGewobagEndpoint_Reachability tests if the Gewobag endpoint is accessible
func TestGewobagEndpoint_Reachability(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := http.NewClient(10 * time.Second)
	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:110.0) Gecko/20100101 Firefox/110.0", // one of the random user agents
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Test basic connectivity to the domain
	resp, err := client.Get(ctx, "https://www.gewobag.de", headers)
	if err != nil {
		t.Fatalf("Gewobag domain unreachable: %v", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		t.Errorf("Gewobag domain returned bad status: %d", resp.StatusCode)
	}
}

// TestGewobagScraper_RealEndpoint tests against the actual Gewobag search endpoint
func TestGewobagScraper_RealEndpoint(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create real dependencies
	httpClient := http.NewClient(30 * time.Second)
	state := store.NewScraperState()
	scraper := common.NewBaseScraper(httpClient, state, "Gewobag", FetchListings)

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	listings, err := scraper.Scrape(ctx)
	if err != nil {
		t.Fatalf("FetchListings failed: %v", err)
	}

	// Validate response structure
	t.Logf("Retrieved %d listings from Gewobag", len(listings))
}

// TestGewobagScraper_HTMLStructure validates the HTML structure we depend on
func TestGewobagScraper_HTMLStructure(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	httpClient := http.NewClient(30 * time.Second)
	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:110.0) Gecko/20100101 Firefox/110.0", // one of the random user agents
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := httpClient.Get(ctx, config.GewobagURL, headers)
	if err != nil {
		t.Fatalf("Failed to fetch Gewobag search page: %v", err)
	}

	// HTTP status code is the primary indicator of success
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		t.Fatalf("Gewobag returned error status: %d (endpoint likely changed)", resp.StatusCode)
	}

	// Ensure we get a reasonable response length
	if len(resp.Body) < 100 {
		t.Error("Response too short, likely not a proper search results page")
	}

	htmlContent := string(resp.Body)

	// Ensure response is HTML
	if !strings.Contains(htmlContent, "<html") && !strings.Contains(htmlContent, "<!DOCTYPE") {
		t.Error("Response doesn't appear to be valid HTML")
	}

	t.Logf("Successfully received HTML response from Gewobag endpoint (status: %d, content length: %d bytes)",
		resp.StatusCode, len(htmlContent))
}
