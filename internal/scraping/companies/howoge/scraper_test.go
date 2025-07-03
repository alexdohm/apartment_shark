package howoge

import (
	"apartmenthunter/internal/config"
	"apartmenthunter/internal/http"
	"apartmenthunter/internal/scraping/common"
	"apartmenthunter/internal/store"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// TestHowogeEndpoint_Reachability tests if the Howoge endpoint is accessible
func TestHowogeEndpoint_Reachability(t *testing.T) {
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
	resp, err := client.Get(ctx, "https://www.howoge.de", headers)
	if err != nil {
		t.Fatalf("Howoge domain unreachable: %v", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		t.Errorf("Howoge domain returned bad status: %d", resp.StatusCode)
	}
}

// TestHowogeScraper_RealEndpoint tests against the actual Howoge API endpoint
func TestHowogeScraper_RealEndpoint(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create real dependencies
	httpClient := http.NewClient(30 * time.Second)
	state := store.NewScraperState()
	scraper := common.NewBaseScraper(httpClient, state, "Howoge", FetchListings)

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	listings, err := scraper.Scrape(ctx)
	if err != nil {
		t.Fatalf("FetchListings failed: %v", err)
	}

	// Validate response structure
	t.Logf("Retrieved %d listings from Howoge", len(listings))
}

// TestHowogeScraper_JSONStructure validates the JSON API structure we depend on
func TestHowogeScraper_JSONStructure(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	httpClient := http.NewClient(30 * time.Second)
	formData := buildFormData()
	headers := map[string]string{
		"User-Agent":   "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:110.0) Gecko/20100101 Firefox/110.0", // one of the random user agents
		"Referer":      "https://www.howoge.de",
		"Origin":       "https://www.howoge.de",
		"Content-Type": "application/x-www-form-urlencoded",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := httpClient.Post(ctx, config.HowogeURL, formData, headers)
	if err != nil {
		t.Fatalf("Failed to fetch Howoge API: %v", err)
	}

	// HTTP status code is the primary indicator of success
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		t.Fatalf("Howoge returned error status: %d (form data likely invalid or endpoint changed)", resp.StatusCode)
	}

	// Ensure we get a reasonable response length
	if len(resp.Body) < 10 {
		t.Error("Response too short, likely not a proper API response")
	}

	// Test JSON parsing
	var response HowogeResponse
	if err := json.Unmarshal(resp.Body, &response); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	t.Logf("Successfully received and parsed JSON response with %d listings", len(response.Results))

	if len(response.Results) == 0 {
		t.Log("No listings in response - (this is not an error)")
	}
}

// TestHowogeScraper_JSONFields validates that required JSON fields exist when listings are present
func TestHowogeScraper_JSONFields(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	httpClient := http.NewClient(30 * time.Second)
	formData := buildFormData()
	headers := map[string]string{
		"User-Agent":   "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:110.0) Gecko/20100101 Firefox/110.0", // one of the random user agents
		"Referer":      "https://www.howoge.de",
		"Origin":       "https://www.howoge.de",
		"Content-Type": "application/x-www-form-urlencoded",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := httpClient.Post(ctx, config.HowogeURL, formData, headers)
	if err != nil {
		t.Fatalf("Failed to fetch Howoge API: %v", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		t.Fatalf("Howoge returned error status: %d", resp.StatusCode)
	}

	var response HowogeResponse
	if err := json.Unmarshal(resp.Body, &response); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	listingCount := len(response.Results)
	t.Logf("Found %d listings in JSON response", listingCount)

	if listingCount == 0 {
		t.Skip("No listings found - cannot validate JSON fields (this is not an error)")
		return
	}

	// Test each listing for required fields
	missingFields := make(map[string]int)
	totalListings := 0

	for i, listing := range response.Results {
		if i >= 3 {
			continue
		}
		totalListings++
		t.Logf("Validating listing %d: ID=%d", i+1, listing.ID)

		// Check required fields that we depend on
		fieldChecks := map[string]interface{}{
			"ID":      listing.ID,
			"Address": listing.Address,
			"Rent":    listing.Rent,
			"Size":    listing.Size,
			"Link":    listing.Link,
		}

		for fieldName, fieldValue := range fieldChecks {
			isEmpty := false
			switch v := fieldValue.(type) {
			case string:
				isEmpty = strings.TrimSpace(v) == ""
			case int:
				isEmpty = v == 0
			case float64:
				isEmpty = v == 0.0
			}

			if isEmpty {
				missingFields[fieldName]++
				t.Logf("Listing %d missing field: %s", i+1, fieldName)
			} else {
				// Log what we found for debugging
				if fieldName == "ID" || fieldName == "Rent" || fieldName == "Size" {
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

// TestHowogeFormData_Validation tests if our form parameters are still valid
func TestHowogeFormData_Validation(t *testing.T) {
	formData := buildFormData()

	if formData == nil {
		t.Fatal("buildFormData returned nil")
	}

	// Validate required fields
	requiredFields := map[string]string{
		"tx_howrealestate_json_list[action]": "immoList",
		"tx_howrealestate_json_list[page]":   "1",
		"tx_howrealestate_json_list[limit]":  "50",
		"tx_howrealestate_json_list[lang]":   "", // Empty is expected
	}

	for field, expectedValue := range requiredFields {
		actualValues, exists := formData[field]
		if !exists {
			t.Errorf("Required form field missing: %s", field)
		} else if len(actualValues) == 0 {
			t.Errorf("Form field %s has no values", field)
		} else if actualValues[0] != expectedValue {
			t.Errorf("Form field %s = %s, want %s", field, actualValues[0], expectedValue)
		}
	}

	t.Logf("Form data validation passed with %d fields", len(formData))
}
