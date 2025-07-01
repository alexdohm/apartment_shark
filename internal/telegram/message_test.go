package telegram

import (
	"strings"
	"testing"
)

func TestBuildHTML_ContainsExpectedElements(t *testing.T) {
	info := &TelegramInfo{
		Address:     "Test",
		Size:        "40",
		Rent:        "400",
		MapLink:     "maps.google.com",
		ListingLink: "google.com",
		Site:        "Test Site",
	}
	result := BuildHTML(info)

	expectedElements := []string{
		"<b>Test Site Listing</b>",
		"<b>Address:</b> Test",
		"<b>Size:</b> 40 m²",
		"<b>Rent:</b> 400 €",
		`<a href="maps.google.com">View Map</a>`,
		`<a href="google.com">View Listing</a>`,
	}

	for _, elem := range expectedElements {
		if !strings.Contains(result, elem) {
			t.Errorf("BuildHTML(): %s expected in %s but not found", elem, result)
		}
	}
}

func TestBuildHTML_HTMLStructure(t *testing.T) {
	info := &TelegramInfo{
		Address:     "Test Addy",
		Size:        "40",
		Rent:        "400",
		MapLink:     "maps.google.com",
		ListingLink: "google.com",
		Site:        "Test Site",
	}
	result := BuildHTML(info)

	if !strings.HasPrefix(result, "<b>Test Site Listing</b>") {
		t.Errorf("BuildHTML() - Header must be bold with site name")
	}

	boldCount := strings.Count(result, "<b>") + strings.Count(result, "</b>")
	if boldCount != 8 { // 4 opening and 4 closing bolds
		t.Errorf("BuildHTML() should have the first 8 bold counts, but got %d", boldCount)
	}

	anchorCount := strings.Count(result, "<a href=") + strings.Count(result, "</a>")
	if anchorCount != 4 { // 2 opening, 2 closing
		t.Errorf("BuildHTML() should have 4 total anchor counts, but got %d", anchorCount)
	}
}

func TestBuildHTML_Fallbacks(t *testing.T) {
	tests := []struct {
		name     string
		info     *TelegramInfo
		expected string
	}{
		{
			name:     "nil info",
			info:     nil,
			expected: "Data not provided",
		},
		{
			name: "all empty fields",
			info: &TelegramInfo{
				Address:     "",
				Size:        "",
				Rent:        "",
				MapLink:     "",
				ListingLink: "",
				Site:        "",
			},
			expected: `<b> Listing</b>

  <b>Address:</b> -
  <b>Size:</b> - m²
  <b>Rent:</b> - €

  <a href="#">View Map</a>
  <a href="#">View Listing</a>`,
		},
		{
			name: "partial data with fallbacks",
			info: &TelegramInfo{
				Address:     "Test Street 123",
				Size:        "",
				Rent:        "800",
				MapLink:     "",
				ListingLink: "https://example.com",
				Site:        "",
			},
			expected: `<b> Listing</b>

  <b>Address:</b> Test Street 123
  <b>Size:</b> - m²
  <b>Rent:</b> 800 €

  <a href="#">View Map</a>
  <a href="https://example.com">View Listing</a>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildHTML(tt.info)
			if result != tt.expected {
				t.Errorf("BuildHTML() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func BenchmarkBuildHTML(b *testing.B) {
	info := &TelegramInfo{
		Address:     "Test Addy",
		Size:        "40",
		Rent:        "400",
		MapLink:     "maps.google.com",
		ListingLink: "google.com",
		Site:        "Gewobag",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BuildHTML(info)
	}
}
