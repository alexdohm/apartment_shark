package gewobag

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

// TestGewobagScraper_CSSSelectors validates that our CSS selectors work when listings exist
func TestGewobagScraper_CSSSelectors(t *testing.T) {
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

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		t.Fatalf("Gewobag returned error status: %d", resp.StatusCode)
	}

	htmlContent := string(resp.Body)
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	// Find all listing articles
	articles := doc.Find("article[id^=post-]")
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
		"tr.angebot-address td address": "address information",
		"tr.angebot-area td":            "area/size information",
		"tr.angebot-kosten td":          "cost information",
		"a.read-more-link":              "listing link",
	}

	articles.EachWithBreak(func(i int, article *goquery.Selection) bool {
		if i >= 3 {
			return false
		}
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
				if selector == "a.read-more-link" {
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
		return true
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
