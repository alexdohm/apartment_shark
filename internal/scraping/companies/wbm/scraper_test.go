package wbm

import (
	"apartmenthunter/internal/config"
	"apartmenthunter/internal/http"
	"apartmenthunter/internal/scraping/common"
	"apartmenthunter/internal/store"
	"bytes"
	"context"
	"github.com/PuerkitoBio/goquery"
	"strings"
	"testing"
	"time"
)

// TestWBMEndpoint_Reachability tests if the WBM endpoint is accessible
func TestWBMEndpoint_Reachability(t *testing.T) {
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
	resp, err := client.Get(ctx, "https://www.wbm.de", headers)
	if err != nil {
		t.Fatalf("WBM domain unreachable: %v", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		t.Errorf("WBM domain returned bad status: %d", resp.StatusCode)
	}
}

// TestWBMScraper_RealEndpoint tests against the actual WBM search endpoint
func TestWBMScraper_RealEndpoint(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create real dependencies
	httpClient := http.NewClient(30 * time.Second)
	state := store.NewScraperState()
	scraper := common.NewBaseScraper(httpClient, state, "WBM", FetchListings)

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	listings, err := scraper.Scrape(ctx)
	if err != nil {
		t.Fatalf("FetchListings failed: %v", err)
	}

	// Validate response structure
	t.Logf("Retrieved %d listings from WBM", len(listings))
}

// TestWBMScraper_HTMLStructure validates the HTML structure we depend on
func TestWBMScraper_HTMLStructure(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	httpClient := http.NewClient(30 * time.Second)
	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:110.0) Gecko/20100101 Firefox/110.0", // one of the random user agents
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := httpClient.Get(ctx, config.WbmURL, headers)
	if err != nil {
		t.Fatalf("Failed to fetch WBM search page: %v", err)
	}

	// HTTP status code is the primary indicator of success
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		t.Fatalf("WBM returned error status: %d (endpoint likely changed)", resp.StatusCode)
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

	t.Logf("Successfully received HTML response from WBM endpoint (status: %d, content length: %d bytes)",
		resp.StatusCode, len(htmlContent))
}

// TestWBMScraper_CSSSelectors validates that our CSS selectors work when listings exist
func TestWBMScraper_CSSSelectors(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	httpClient := http.NewClient(30 * time.Second)
	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:110.0) Gecko/20100101 Firefox/110.0", // one of the random user agents
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := httpClient.Get(ctx, config.WbmURL, headers)
	if err != nil {
		t.Fatalf("Failed to fetch WBM search page: %v", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		t.Fatalf("WBM returned error status: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	// Find all listing items
	items := doc.Find("div.row.openimmo-search-list-item")
	itemCount := items.Length()

	t.Logf("Found %d listing items", itemCount)

	if itemCount == 0 {
		t.Skip("No listings found - cannot validate CSS selectors (this is not an error)")
		return
	}

	// Test each item for required selectors
	missingSelectors := make(map[string]int)
	totalItems := 0
	// Define the selectors we depend on
	requiredSelectors := map[string]string{
		"div.address": "address information",
		"div.main-property-value.main-property-rent": "rent information",
		"div.main-property-value.main-property-size": "size information",
		"div.btn-holder a":                           "listing link",
	}

	items.EachWithBreak(func(i int, item *goquery.Selection) bool {
		if i >= 3 {
			return false
		}
		totalItems++
		itemID, _ := item.Attr("data-id")
		t.Logf("Validating item %d: data-id=%s", i+1, itemID)

		for selector, description := range requiredSelectors {
			element := item.Find(selector)
			if element.Length() == 0 {
				missingSelectors[selector]++
				t.Logf("Item %s missing selector: %s (%s)", itemID, selector, description)
			} else {
				// Log what we found for debugging
				text := strings.TrimSpace(element.Text())
				if selector == "div.btn-holder a" {
					if href, exists := element.Attr("href"); exists {
						t.Logf("Item %s %s: href=%s", itemID, description, href)
					} else {
						t.Logf("Item %s %s: no href attribute", itemID, description)
					}
				} else {
					t.Logf("Item %s %s: '%s'", itemID, description, text)
				}
			}
		}

		// Test data-id attribute specifically
		if itemID == "" {
			missingSelectors["data-id attribute"]++
			t.Logf("Item %d missing data-id attribute", i+1)
		}
		return true
	})

	// Report results
	if len(missingSelectors) > 0 {
		t.Errorf("CSS selector validation failed:")
		for selector, count := range missingSelectors {
			percentage := float64(count) / float64(totalItems) * 100
			t.Errorf("  - %s missing in %d/%d items (%.1f%%)", selector, count, totalItems, percentage)
		}
	} else {
		t.Logf("SUCCESS: All required CSS selectors found in all %d items", totalItems)
	}
}
