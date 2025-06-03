package telegram

import (
	"strings"
	"testing"
)

func TestBuildHTML(t *testing.T) {
	tests := []struct {
		name     string
		info     *TelegramInfo
		site     string
		expected string
	}{
		{
			name: "standard apartment listing",
			info: &TelegramInfo{
				Address:     "Kienitzer Str. 24, 12053 Berlin",
				Size:        "65",
				Rent:        "800",
				MapLink:     "https://www.google.com/maps/place/Kienitzer+Str.+24,+12053+Berlin/",
				ListingLink: "google.com",
			},
			site: "StadtUndLand",
			expected: `<b>StadtUndLand Listing</b>

<b>Address:</b> Kienitzer Str. 24, 12053 Berlin
<b>Size:</b> 65 m²
<b>Rent:</b> 800 €

<a href="https://www.google.com/maps/place/Kienitzer+Str.+24,+12053+Berlin/">View Map</a>

<a href="google.com">View Listing</a>`,
		},
		{
			name: "empty values",
			info: &TelegramInfo{
				Address:     "",
				Size:        "",
				Rent:        "",
				MapLink:     "",
				ListingLink: "",
			},
			site: "Test",
			expected: `<b>Test Listing</b>

<b>Address:</b> 
<b>Size:</b>  m²
<b>Rent:</b>  €

<a href="">View Map</a>

<a href="">View Listing</a>`,
		},
		{
			name: "special characters in address",
			info: &TelegramInfo{
				Address:     "Käthe Str. 24, 12053 Berlin",
				Size:        "65",
				Rent:        "800",
				MapLink:     "https://www.google.com/maps/place/Kienitzer+Str.+24,+12053+Berlin/",
				ListingLink: "google.com",
			},
			site: "Gewobag",
			expected: `<b>Gewobag Listing</b>

<b>Address:</b> Käthe Str. 24, 12053 Berlin
<b>Size:</b> 65 m²
<b>Rent:</b> 800 €

<a href="https://www.google.com/maps/place/Kienitzer+Str.+24,+12053+Berlin/">View Map</a>

<a href="google.com">View Listing</a>`,
		},
		{
			name:     "nil info struct",
			info:     nil, // this will cause site to panic
			site:     "Test",
			expected: ``,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.info == nil {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("BuildHTML() should panic with nil info, but didnt")
					}
				}()
				BuildHTML(tt.info, tt.site)
				return
			}

			result := BuildHTML(tt.info, tt.site)
			if result != tt.expected {
				t.Errorf("BuildHTML() mismatch:\nGot\n%s\n\nExpected\n%s\n", result, tt.expected)
			}
		})
	}
}

func TestBuildHTML_ContainsExpectedElements(t *testing.T) {
	info := &TelegramInfo{
		Address:     "Test",
		Size:        "40",
		Rent:        "400",
		MapLink:     "maps.google.com",
		ListingLink: "google.com",
	}
	site := "Test Site"
	result := BuildHTML(info, site)

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
	}
	site := "Test Site"
	result := BuildHTML(info, site)

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

func BenchmarkBuildHTML(b *testing.B) {
	info := &TelegramInfo{
		Address:     "Test Addy",
		Size:        "40",
		Rent:        "400",
		MapLink:     "maps.google.com",
		ListingLink: "google.com",
	}
	site := "Gewobag"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BuildHTML(info, site)
	}

}
