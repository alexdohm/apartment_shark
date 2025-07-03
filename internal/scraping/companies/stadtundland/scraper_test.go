package stadtundland

import (
	"apartmenthunter/internal/config"
	"apartmenthunter/internal/http"
	"apartmenthunter/internal/scraping/common"
	"apartmenthunter/internal/store"
	"context"
	"encoding/json"
	"testing"
	"time"
)

// TestStadtUndLandEndpoint_Reachability tests if the Stadt und Land endpoint is accessible
func TestStadtUndLandEndpoint_Reachability(t *testing.T) {
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
	resp, err := client.Get(ctx, "https://stadtundland.de", headers)
	if err != nil {
		t.Fatalf("Stadt und Land domain unreachable: %v", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		t.Errorf("Stadt und Land domain returned bad status: %d", resp.StatusCode)
	}
}

// TestStadtUndLandScraper_RealEndpoint tests against the actual Stadt und Land API endpoint
func TestStadtUndLandScraper_RealEndpoint(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create real dependencies
	httpClient := http.NewClient(30 * time.Second)
	state := store.NewScraperState()
	scraper := common.NewBaseScraper(httpClient, state, "Stadt Und Land", FetchListings)

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	listings, err := scraper.Scrape(ctx)
	if err != nil {
		t.Fatalf("FetchListings failed: %v", err)
	}

	// Validate response structure
	t.Logf("Retrieved %d listings from Stadt und Land", len(listings))
}

// TestStadtUndLandScraper_JSONStructure validates the JSON API structure we depend on
func TestStadtUndLandScraper_JSONStructure(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	httpClient := http.NewClient(30 * time.Second)
	jsonData := buildFormData()
	headers := map[string]string{
		"User-Agent":   "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:110.0) Gecko/20100101 Firefox/110.0", // one of the random user agents
		"Content-Type": "application/json",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := httpClient.PostJSON(ctx, config.StadtUndLandURL, jsonData, headers)
	if err != nil {
		t.Fatalf("Failed to fetch Stadt und Land API: %v", err)
	}

	// HTTP status code is the primary indicator of success
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		t.Fatalf("Stadt und Land returned error status: %d (JSON data likely invalid or endpoint changed)", resp.StatusCode)
	}

	// Ensure we get a reasonable response length
	if len(resp.Body) < 10 {
		t.Error("Response too short, likely not a proper API response")
	}

	// Test JSON parsing
	var response StadtUndLandResponse
	if err := json.Unmarshal(resp.Body, &response); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	t.Logf("Successfully received and parsed JSON response with %d listings", len(response.Listings))

	if len(response.Listings) == 0 {
		t.Log("No listings in response - cannot validate JSON structure (this is not an error)")
	}
}

// TestStadtUndLandScraper_JSONFields validates that required JSON fields exist when listings are present
func TestStadtUndLandScraper_JSONFields(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	httpClient := http.NewClient(30 * time.Second)
	jsonData := buildFormData()
	headers := map[string]string{
		"User-Agent":   "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:110.0) Gecko/20100101 Firefox/110.0", // one of the random user agents
		"Content-Type": "application/json",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := httpClient.PostJSON(ctx, config.StadtUndLandURL, jsonData, headers)
	if err != nil {
		t.Fatalf("Failed to fetch Stadt und Land API: %v", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		t.Fatalf("Stadt und Land returned error status: %d", resp.StatusCode)
	}

	var response StadtUndLandResponse
	if err := json.Unmarshal(resp.Body, &response); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	listingCount := len(response.Listings)
	t.Logf("Found %d listings in JSON response", listingCount)

	if listingCount == 0 {
		t.Skip("No listings found - cannot validate JSON fields (this is not an error)")
		return
	}

	// Test each listing for required fields
	missingFields := make(map[string]int)
	totalListings := 0

	for i, listing := range response.Listings {
		if i >= 3 {
			continue
		}
		totalListings++
		t.Logf("Validating listing %d: ID=%s", i+1, listing.Details.Id)

		// Check required fields that we depend on
		fieldChecks := map[string]interface{}{
			"Details.Id":          listing.Details.Id,
			"Details.Area":        listing.Details.Area,
			"Costs.Rent":          listing.Costs.Rent,
			"Address.Street":      listing.Address.Street,
			"Address.HouseNumber": listing.Address.HouseNumber,
			"Address.PostalCode":  listing.Address.PostalCode,
			"Address.City":        listing.Address.City,
		}

		for fieldName, fieldValue := range fieldChecks {
			if fieldValue == "" {
				missingFields[fieldName]++
				t.Logf("Listing %d missing field: %s", i+1, fieldName)
			} else {
				// Log what we found for debugging
				if fieldName == "Details.Id" || fieldName == "Details.Area" || fieldName == "Costs.Rent" {
					t.Logf("Listing %d %s: '%v'", i+1, fieldName, fieldValue)
				}
			}
		}
	}

	// Report results
	if len(missingFields) > 0 {
		t.Errorf("JSON field validation failed:")
		for fieldName, count := range missingFields {
			percentage := float64(count) / float64(totalListings) * 100
			t.Errorf("  - %s missing in %d/%d listings (%.1f%%)", fieldName, count, totalListings, percentage)

		}
	} else {
		t.Logf("SUCCESS: All required JSON fields found in all %d listings", totalListings)
	}
}
