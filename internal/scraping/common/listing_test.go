package common

import (
	"apartmenthunter/internal/telegram"
	"apartmenthunter/internal/users"
	"strings"
	"testing"
)

// TestListing_ToTelegramInfo tests the Listing to TelegramInfo conversion
func TestListing_ToTelegramInfo(t *testing.T) {
	tests := []struct {
		name     string
		listing  Listing
		expected *telegram.TelegramInfo
	}{
		{
			name: "complete listing with normal characters",
			listing: Listing{
				ID:      "12345",
				Company: "TestCompany",
				Price:   "800",
				Size:    "45.5",
				Address: "Musterstraße 1, 10115 Berlin",
				URL:     "https://example.com/listing/12345",
			},
			expected: &telegram.TelegramInfo{
				Address:     "Musterstraße 1, 10115 Berlin",
				Size:        "45.5",
				Rent:        "800",
				MapLink:     "https://www.google.com/maps/search/?api=1&query=Musterstra%C3%9Fe+1%2C+10115+Berlin",
				ListingLink: "https://example.com/listing/12345",
				Site:        "TestCompany",
			},
		},
		{
			name: "address with special characters and spaces",
			listing: Listing{
				ID:      "67890",
				Company: "Ümlaut Company",
				Price:   "1200",
				Size:    "60.0",
				Address: "Königstraße 123, Berlin-Mitte",
				URL:     "https://example.com/listing/67890",
			},
			expected: &telegram.TelegramInfo{
				Address:     "Königstraße 123, Berlin-Mitte",
				Size:        "60.0",
				Rent:        "1200",
				MapLink:     "https://www.google.com/maps/search/?api=1&query=K%C3%B6nigstra%C3%9Fe+123%2C+Berlin-Mitte",
				ListingLink: "https://example.com/listing/67890",
				Site:        "Ümlaut Company",
			},
		},
		{
			name: "address with symbols and punctuation",
			listing: Listing{
				ID:      "special",
				Company: "Test & Co.",
				Price:   "950",
				Size:    "42.5",
				Address: "Straße der 17. Juni 135, 10623 Berlin",
				URL:     "https://test.com/apt?id=special&ref=search",
			},
			expected: &telegram.TelegramInfo{
				Address:     "Straße der 17. Juni 135, 10623 Berlin",
				Size:        "42.5",
				Rent:        "950",
				MapLink:     "https://www.google.com/maps/search/?api=1&query=Stra%C3%9Fe+der+17.+Juni+135%2C+10623+Berlin",
				ListingLink: "https://test.com/apt?id=special&ref=search",
				Site:        "Test & Co.",
			},
		},
		{
			name: "empty fields",
			listing: Listing{
				ID:      "",
				Company: "",
				Price:   "",
				Size:    "",
				Address: "",
				URL:     "",
			},
			expected: &telegram.TelegramInfo{
				Address:     "",
				Size:        "",
				Rent:        "",
				MapLink:     "https://www.google.com/maps/search/?api=1&query=",
				ListingLink: "",
				Site:        "",
			},
		},
		{
			name: "address with only spaces",
			listing: Listing{
				ID:      "spaces",
				Company: "SpaceTest",
				Price:   "700",
				Size:    "30",
				Address: "   ",
				URL:     "https://example.com",
			},
			expected: &telegram.TelegramInfo{
				Address:     "   ",
				Size:        "30",
				Rent:        "700",
				MapLink:     "https://www.google.com/maps/search/?api=1&query=+++",
				ListingLink: "https://example.com",
				Site:        "SpaceTest",
			},
		},
		{
			name: "long address with multiple special characters",
			listing: Listing{
				ID:      "long",
				Company: "LongAddress Inc.",
				Price:   "1500",
				Size:    "75.5",
				Address: "Am Köllnischen Park 1-3, Apartment 4A/B, 10179 Berlin-Mitte, Germany",
				URL:     "https://example.com/long-listing",
			},
			expected: &telegram.TelegramInfo{
				Address:     "Am Köllnischen Park 1-3, Apartment 4A/B, 10179 Berlin-Mitte, Germany",
				Size:        "75.5",
				Rent:        "1500",
				MapLink:     "https://www.google.com/maps/search/?api=1&query=Am+K%C3%B6llnischen+Park+1-3%2C+Apartment+4A%2FB%2C+10179+Berlin-Mitte%2C+Germany",
				ListingLink: "https://example.com/long-listing",
				Site:        "LongAddress Inc.",
			},
		},
		{
			name: "numeric and decimal values",
			listing: Listing{
				ID:      "123456789",
				Company: "Numeric Co. 2023",
				Price:   "999.50",
				Size:    "42.75",
				Address: "Teststraße 42, 12345 Berlin",
				URL:     "https://numeric.example.com/listing/999",
			},
			expected: &telegram.TelegramInfo{
				Address:     "Teststraße 42, 12345 Berlin",
				Size:        "42.75",
				Rent:        "999.50",
				MapLink:     "https://www.google.com/maps/search/?api=1&query=Teststra%C3%9Fe+42%2C+12345+Berlin",
				ListingLink: "https://numeric.example.com/listing/999",
				Site:        "Numeric Co. 2023",
			},
		},
		{
			name: "address with forward slashes and parentheses",
			listing: Listing{
				ID:      "complex",
				Company: "Complex/Address (Test)",
				Price:   "850",
				Size:    "50",
				Address: "Straße/Gasse 12 (Hinterhof), 10115 Berlin",
				URL:     "https://example.com/complex?search=(test)",
			},
			expected: &telegram.TelegramInfo{
				Address:     "Straße/Gasse 12 (Hinterhof), 10115 Berlin",
				Size:        "50",
				Rent:        "850",
				MapLink:     "https://www.google.com/maps/search/?api=1&query=Stra%C3%9Fe%2FGasse+12+%28Hinterhof%29%2C+10115+Berlin",
				ListingLink: "https://example.com/complex?search=(test)",
				Site:        "Complex/Address (Test)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.listing.ToTelegramInfo()

			// Check all fields
			if result.Address != tt.expected.Address {
				t.Errorf("Address = %q, want %q", result.Address, tt.expected.Address)
			}
			if result.Size != tt.expected.Size {
				t.Errorf("Size = %q, want %q", result.Size, tt.expected.Size)
			}
			if result.Rent != tt.expected.Rent {
				t.Errorf("Rent = %q, want %q", result.Rent, tt.expected.Rent)
			}
			if result.MapLink != tt.expected.MapLink {
				t.Errorf("MapLink = %q, want %q", result.MapLink, tt.expected.MapLink)
			}
			if result.ListingLink != tt.expected.ListingLink {
				t.Errorf("ListingLink = %q, want %q", result.ListingLink, tt.expected.ListingLink)
			}
			if result.Site != tt.expected.Site {
				t.Errorf("Site = %q, want %q", result.Site, tt.expected.Site)
			}
		})
	}
}

// TestListing_ToTelegramInfo_FieldMapping ensures all fields are mapped correctly
func TestListing_ToTelegramInfo_FieldMapping(t *testing.T) {
	listing := Listing{
		ID:      "field-test",
		Company: "Field Test Company",
		Price:   "1234.56",
		Size:    "67.89",
		Address: "Field Test Address",
		URL:     "https://field-test.example.com",
	}

	result := listing.ToTelegramInfo()

	// Verify field mapping
	if result.Site != listing.Company {
		t.Errorf("Site field mapping: got %q, want %q", result.Site, listing.Company)
	}
	if result.Rent != listing.Price {
		t.Errorf("Rent field mapping: got %q, want %q", result.Rent, listing.Price)
	}
	if result.Size != listing.Size {
		t.Errorf("Size field mapping: got %q, want %q", result.Size, listing.Size)
	}
	if result.Address != listing.Address {
		t.Errorf("Address field mapping: got %q, want %q", result.Address, listing.Address)
	}
	if result.ListingLink != listing.URL {
		t.Errorf("ListingLink field mapping: got %q, want %q", result.ListingLink, listing.URL)
	}

	// Verify MapLink contains the address
	if !strings.Contains(result.MapLink, "Field+Test+Address") {
		t.Errorf("MapLink should contain encoded address, got %q", result.MapLink)
	}
}

// TestListing_MatchUserConfig tests the user configuration matching functionality
func TestListing_MatchUserConfig(t *testing.T) {
	tests := []struct {
		name       string
		listing    Listing
		userConfig users.UserConfig
		expected   bool
	}{
		{
			name: "perfect match - all criteria met",
			listing: Listing{
				ZipCode:     "12043",
				WbsRequired: true,
				Price:       "800",
				Size:        "60",
			},
			userConfig: users.UserConfig{
				ZipCodes:    []string{"12043", "12045"},
				WbsRequired: true,
				MinPrice:    500,
				MaxPrice:    1000,
				MinSqm:      50,
				MaxSqm:      70,
			},
			expected: true,
		},
		{
			name: "zip code mismatch",
			listing: Listing{
				ZipCode:     "10115",
				WbsRequired: true,
				Price:       "800",
				Size:        "60",
			},
			userConfig: users.UserConfig{
				ZipCodes:    []string{"12043", "12045"},
				WbsRequired: true,
				MinPrice:    500,
				MaxPrice:    1000,
				MinSqm:      50,
				MaxSqm:      70,
			},
			expected: false,
		},
		{
			name: "WBS requirement not met",
			listing: Listing{
				ZipCode:     "12043",
				WbsRequired: false,
				Price:       "800",
				Size:        "60",
			},
			userConfig: users.UserConfig{
				ZipCodes:    []string{"12043"},
				WbsRequired: true,
				MinPrice:    500,
				MaxPrice:    1000,
				MinSqm:      50,
				MaxSqm:      70,
			},
			expected: false,
		},
		{
			name: "WBS not required by user, listing has WBS",
			listing: Listing{
				ZipCode:     "12043",
				WbsRequired: true,
				Price:       "800",
				Size:        "60",
			},
			userConfig: users.UserConfig{
				ZipCodes:    []string{"12043"},
				WbsRequired: false,
				MinPrice:    500,
				MaxPrice:    1000,
				MinSqm:      50,
				MaxSqm:      70,
			},
			expected: true,
		},
		{
			name: "price too low",
			listing: Listing{
				ZipCode:     "12043",
				WbsRequired: false,
				Price:       "400",
				Size:        "60",
			},
			userConfig: users.UserConfig{
				ZipCodes:    []string{"12043"},
				WbsRequired: false,
				MinPrice:    500,
				MaxPrice:    1000,
				MinSqm:      50,
				MaxSqm:      70,
			},
			expected: false,
		},
		{
			name: "price too high",
			listing: Listing{
				ZipCode:     "12043",
				WbsRequired: false,
				Price:       "1200",
				Size:        "60",
			},
			userConfig: users.UserConfig{
				ZipCodes:    []string{"12043"},
				WbsRequired: false,
				MinPrice:    500,
				MaxPrice:    1000,
				MinSqm:      50,
				MaxSqm:      70,
			},
			expected: false,
		},
		{
			name: "size too small",
			listing: Listing{
				ZipCode:     "12043",
				WbsRequired: false,
				Price:       "800",
				Size:        "40",
			},
			userConfig: users.UserConfig{
				ZipCodes:    []string{"12043"},
				WbsRequired: false,
				MinPrice:    500,
				MaxPrice:    1000,
				MinSqm:      50,
				MaxSqm:      70,
			},
			expected: false,
		},
		{
			name: "size too large",
			listing: Listing{
				ZipCode:     "12043",
				WbsRequired: false,
				Price:       "800",
				Size:        "80",
			},
			userConfig: users.UserConfig{
				ZipCodes:    []string{"12043"},
				WbsRequired: false,
				MinPrice:    500,
				MaxPrice:    1000,
				MinSqm:      50,
				MaxSqm:      70,
			},
			expected: false,
		},
		{
			name: "no zip code restrictions",
			listing: Listing{
				ZipCode:     "99999",
				WbsRequired: false,
				Price:       "800",
				Size:        "60",
			},
			userConfig: users.UserConfig{
				ZipCodes:    []string{},
				WbsRequired: false,
				MinPrice:    500,
				MaxPrice:    1000,
				MinSqm:      50,
				MaxSqm:      70,
			},
			expected: true,
		},
		{
			name: "no price restrictions",
			listing: Listing{
				ZipCode:     "12043",
				WbsRequired: false,
				Price:       "2000",
				Size:        "60",
			},
			userConfig: users.UserConfig{
				ZipCodes:    []string{"12043"},
				WbsRequired: false,
				MinPrice:    0,
				MaxPrice:    0,
				MinSqm:      50,
				MaxSqm:      70,
			},
			expected: true,
		},
		{
			name: "no size restrictions",
			listing: Listing{
				ZipCode:     "12043",
				WbsRequired: false,
				Price:       "800",
				Size:        "200",
			},
			userConfig: users.UserConfig{
				ZipCodes:    []string{"12043"},
				WbsRequired: false,
				MinPrice:    500,
				MaxPrice:    1000,
				MinSqm:      0,
				MaxSqm:      0,
			},
			expected: true,
		},
		{
			name: "price with decimal values",
			listing: Listing{
				ZipCode:     "12043",
				WbsRequired: false,
				Price:       "799.50",
				Size:        "60.5",
			},
			userConfig: users.UserConfig{
				ZipCodes:    []string{"12043"},
				WbsRequired: false,
				MinPrice:    500,
				MaxPrice:    1000,
				MinSqm:      50,
				MaxSqm:      70,
			},
			expected: true,
		},
		{
			name: "price with currency symbol",
			listing: Listing{
				ZipCode:     "12043",
				WbsRequired: false,
				Price:       "€800",
				Size:        "60 qm",
			},
			userConfig: users.UserConfig{
				ZipCodes:    []string{"12043"},
				WbsRequired: false,
				MinPrice:    500,
				MaxPrice:    1000,
				MinSqm:      50,
				MaxSqm:      70,
			},
			expected: true,
		},
		{
			name: "unparseable price",
			listing: Listing{
				ZipCode:     "12043",
				WbsRequired: false,
				Price:       "auf Anfrage",
				Size:        "60",
			},
			userConfig: users.UserConfig{
				ZipCodes:    []string{"12043"},
				WbsRequired: false,
				MinPrice:    500,
				MaxPrice:    1000,
				MinSqm:      50,
				MaxSqm:      70,
			},
			expected: false,
		},
		{
			name: "unparseable size",
			listing: Listing{
				ZipCode:     "12043",
				WbsRequired: false,
				Price:       "800",
				Size:        "variabel",
			},
			userConfig: users.UserConfig{
				ZipCodes:    []string{"12043"},
				WbsRequired: false,
				MinPrice:    500,
				MaxPrice:    1000,
				MinSqm:      50,
				MaxSqm:      70,
			},
			expected: false,
		},
		{
			name: "only minimum price set",
			listing: Listing{
				ZipCode:     "12043",
				WbsRequired: false,
				Price:       "600",
				Size:        "60",
			},
			userConfig: users.UserConfig{
				ZipCodes:    []string{"12043"},
				WbsRequired: false,
				MinPrice:    500,
				MaxPrice:    0,
				MinSqm:      50,
				MaxSqm:      70,
			},
			expected: true,
		},
		{
			name: "only maximum price set",
			listing: Listing{
				ZipCode:     "12043",
				WbsRequired: false,
				Price:       "800",
				Size:        "60",
			},
			userConfig: users.UserConfig{
				ZipCodes:    []string{"12043"},
				WbsRequired: false,
				MinPrice:    0,
				MaxPrice:    1000,
				MinSqm:      50,
				MaxSqm:      70,
			},
			expected: true,
		},
		{
			name: "only minimum size set",
			listing: Listing{
				ZipCode:     "12043",
				WbsRequired: false,
				Price:       "800",
				Size:        "60",
			},
			userConfig: users.UserConfig{
				ZipCodes:    []string{"12043"},
				WbsRequired: false,
				MinPrice:    500,
				MaxPrice:    1000,
				MinSqm:      50,
				MaxSqm:      0,
			},
			expected: true,
		},
		{
			name: "only maximum size set",
			listing: Listing{
				ZipCode:     "12043",
				WbsRequired: false,
				Price:       "800",
				Size:        "60",
			},
			userConfig: users.UserConfig{
				ZipCodes:    []string{"12043"},
				WbsRequired: false,
				MinPrice:    500,
				MaxPrice:    1000,
				MinSqm:      0,
				MaxSqm:      70,
			},
			expected: true,
		},
		{
			name: "empty zip code in listing",
			listing: Listing{
				ZipCode:     "",
				WbsRequired: false,
				Price:       "800",
				Size:        "60",
			},
			userConfig: users.UserConfig{
				ZipCodes:    []string{"12043"},
				WbsRequired: false,
				MinPrice:    500,
				MaxPrice:    1000,
				MinSqm:      50,
				MaxSqm:      70,
			},
			expected: false,
		},
		{
			name: "complex price string with multiple numbers",
			listing: Listing{
				ZipCode:     "12043",
				WbsRequired: false,
				Price:       "Kaltmiete: 750€, Nebenkosten: 150€",
				Size:        "60.5 qm",
			},
			userConfig: users.UserConfig{
				ZipCodes:    []string{"12043"},
				WbsRequired: false,
				MinPrice:    500,
				MaxPrice:    1000,
				MinSqm:      50,
				MaxSqm:      70,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.listing.MatchUserConfig(&tt.userConfig)
			if result != tt.expected {
				t.Errorf("MatchUserConfig() = %v, want %v", result, tt.expected)
			}
		})
	}
}
