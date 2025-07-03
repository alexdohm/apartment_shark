package dewego

import (
	"apartmenthunter/internal/config"
	"apartmenthunter/internal/http"
	"apartmenthunter/internal/scraping/common"
	"apartmenthunter/internal/store"
	"context"
	"github.com/PuerkitoBio/goquery"
	"strings"
	"testing"
	"time"
)

// integration and contract tests for dewego
func TestDewegoEndpoint_Reachability(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := http.NewClient(10 * time.Second)
	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:110.0) Gecko/20100101 Firefox/110.0", // one of the random user agents
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Test basic connectivity to domain
	resp, err := client.Get(ctx, "https://www.degewo.de", headers)
	if err != nil {
		t.Fatalf("Dewego domain unreachable: %v", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		t.Errorf("Dewego domain returned bad status: %d", resp.StatusCode)
	}
}

// TestDewegoScraper_RealEndpoint tests against the actual Dewego search endpoint
func TestDewegoScraper_RealEndpoint(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create real dependencies
	httpClient := http.NewClient(30 * time.Second)
	state := store.NewScraperState()
	scraper := common.NewBaseScraper(httpClient, state, "Dewego", FetchListings)

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	listings, err := scraper.Scrape(ctx)
	if err != nil {
		t.Fatalf("FetchListings failed: %v", err)
	}

	// Validate response structure
	t.Logf("Retrieved %d listings from Dewego", len(listings))
}

// TestDewegoScraper_DataStructure validates that our form data is accepted by the endpoint
func TestDewegoScraper_DataStructure(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	httpClient := http.NewClient(30 * time.Second)
	formData := buildFormData()
	headers := map[string]string{
		"User-Agent":   "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:110.0) Gecko/20100101 Firefox/110.0", // one of the random user agents
		"Content-Type": "application/x-www-form-urlencoded",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := httpClient.Post(ctx, config.DewegoURL, formData, headers)
	if err != nil {
		t.Fatalf("Failed to fetch Dewego search page: %v", err)
	}

	// Test 1: HTTP response code indicates success
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		t.Fatalf("Dewego returned error status: %d (form data likely invalid or endpoint changed)", resp.StatusCode)
	}

	htmlContent := string(resp.Body)

	// Test 2: Ensure we get a reasonable response length (not just a redirect or empty page)
	if len(htmlContent) < 100 {
		t.Error("Response too short, likely not a proper search results page")
	}

	// Test 3: Ensure response is HTML (not JSON error or plain text)
	if !strings.Contains(htmlContent, "<html") && !strings.Contains(htmlContent, "<!DOCTYPE") {
		t.Error("Response doesn't appear to be valid HTML")
	}

	t.Logf("Successfully submitted form data to Dewego endpoint (status: %d, content length: %d bytes)",
		resp.StatusCode, len(htmlContent))
}

// TestDewegoScraper_CSSSelectors validates that our CSS selectors work when listings exist
func TestDewegoScraper_CSSSelectors(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	httpClient := http.NewClient(30 * time.Second)
	formData := buildFormData()
	headers := map[string]string{
		"User-Agent":   "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:110.0) Gecko/20100101 Firefox/110.0", // one of the random user agents
		"Content-Type": "application/x-www-form-urlencoded",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := httpClient.Post(ctx, config.DewegoURL, formData, headers)
	if err != nil {
		t.Fatalf("Failed to fetch Dewego search page: %v", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		t.Fatalf("Dewego returned error status: %d", resp.StatusCode)
	}

	htmlContent := string(resp.Body)
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	// Find all listing articles
	articles := doc.Find("article[id^=immobilie-list-item]")
	articleCount := articles.Length()

	t.Logf("Found %d listing articles", articleCount)

	if articleCount == 0 {
		t.Skip("No listings found - cannot validate CSS selectors (this is not an error)")
		return
	}

	// Test each article for required selectors
	missingSelectors := make(map[string]int)
	totalArticles := 0
	// Define the selectors we depend on
	requiredSelectors := map[string]string{
		"span.article__meta": "address metadata",
		"ul.article__properties li:nth-child(2) span.text": "size information",
		"div.article__price-tag span.price":                "price information",
		"a[target=_blank]":                                 "listing link",
	}

	articles.Each(func(i int, article *goquery.Selection) {
		totalArticles++
		articleID, _ := article.Attr("id")
		t.Logf("Validating article %d: %s", i+1, articleID)

		for selector, description := range requiredSelectors {
			element := article.Find(selector)
			if element.Length() == 0 {
				missingSelectors[selector]++
				t.Logf("Article %s missing selector: %s (%s)", articleID, selector, description)
			} else {
				// Log what we found for debugging
				text := strings.TrimSpace(element.Text())
				if selector == "a[target=_blank]" {
					if href, exists := element.Attr("href"); exists {
						t.Logf("Article %s %s: href=%s", articleID, description, href)
					} else {
						t.Logf("Article %s %s: no href attribute", articleID, description)
					}
				} else {
					t.Logf("Article %s %s: '%s'", articleID, description, text)
				}
			}
		}
	})

	// Report results
	if len(missingSelectors) > 0 {
		t.Errorf("CSS selector validation failed:")
		for selector, count := range missingSelectors {
			percentage := float64(count) / float64(totalArticles) * 100
			t.Errorf("  - %s missing in %d/%d articles (%.1f%%)", selector, count, totalArticles, percentage)
		}
	} else {
		t.Logf("SUCCESS: All required CSS selectors found in all %d articles", totalArticles)
	}
}
