package common

import "testing"

// TestFilterWBSString tests the WBS string filtering functionality
func TestFilterWBSString(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		expected bool
	}{
		// Positive cases - listings that require WBS
		{
			name:     "explicit wbs mention",
			title:    "Schöne Wohnung mit WBS",
			expected: true,
		},
		{
			name:     "wbs required",
			title:    "2-Zimmer, WBS erforderlich",
			expected: true,
		},
		{
			name:     "wbs pflicht",
			title:    "Apartment - WBS Pflicht",
			expected: true,
		},
		{
			name:     "nur mit wbs",
			title:    "Nur mit WBS verfügbar",
			expected: true,
		},
		{
			name:     "mit wbs",
			title:    "Mit WBS zu vermieten",
			expected: true,
		},
		{
			name:     "full wohnberechtigungsschein",
			title:    "Wohnberechtigungsschein erforderlich",
			expected: true,
		},
		{
			name:     "wbs at end of title",
			title:    "3-Zimmer Wohnung WBS",
			expected: true,
		},
		{
			name:     "mixed case wbs",
			title:    "Apartment WbS erforderlich",
			expected: true,
		},
		{
			name:     "wbs in middle of sentence",
			title:    "Schöne Wohnung, WBS notwendig, zentral gelegen",
			expected: true,
		},

		// Negative cases - listings without WBS requirement
		{
			name:     "ohne wbs",
			title:    "Schöne Wohnung ohne WBS",
			expected: false,
		},
		{
			name:     "ohne wohnberechtigungsschein",
			title:    "Ohne Wohnberechtigungsschein",
			expected: false,
		},
		{
			name:     "wbs-frei",
			title:    "WBS-frei zu vermieten",
			expected: false,
		},
		{
			name:     "wbs frei",
			title:    "WBS frei verfügbar",
			expected: false,
		},
		{
			name:     "o. wbs",
			title:    "Apartment o. WBS",
			expected: false,
		},
		{
			name:     "kein wbs",
			title:    "Kein WBS erforderlich",
			expected: false,
		},
		{
			name:     "keine wbs",
			title:    "Keine WBS notwendig",
			expected: false,
		},
		{
			name:     "wbs nicht erforderlich",
			title:    "WBS nicht erforderlich",
			expected: false,
		},
		{
			name:     "ohne wohnberechtigung",
			title:    "Ohne Wohnberechtigung",
			expected: false,
		},

		// Edge cases
		{
			name:     "no wbs mention",
			title:    "Schöne 2-Zimmer Wohnung in Berlin",
			expected: false,
		},
		{
			name:     "empty string",
			title:    "",
			expected: false,
		},
		{
			name:     "only whitespace",
			title:    "   ",
			expected: false,
		},
		{
			name:     "wbs as part of other word",
			title:    "Wohnungsbaugesellschaft vermietet",
			expected: false, // "wbs" is part of "Wohnungsbaugesellschaft"
		},
		{
			name:     "negative overrides positive",
			title:    "WBS erforderlich, aber ohne WBS möglich",
			expected: false, // Negative terms should override positive
		},
		{
			name:     "multiple wbs mentions",
			title:    "WBS WBS WBS erforderlich",
			expected: true,
		},
		{
			name:     "wbs with special characters",
			title:    "WBS: erforderlich!",
			expected: true,
		},
		{
			name:     "mixed language",
			title:    "Beautiful apartment, WBS required",
			expected: true,
		},
		{
			name:     "all caps",
			title:    "APARTMENT MIT WBS",
			expected: true,
		},
		{
			name:     "mixed case negative",
			title:    "OHNE WBS verfügbar",
			expected: false,
		},
		{
			name:     "partial word match",
			title:    "wbsimilar word",
			expected: true, // Current implementation would match this
		},
		{
			name:     "umlauts and special chars",
			title:    "Schöne Wöhnung, WBS erforderlich",
			expected: true,
		},
		{
			name:     "long complex title",
			title:    "Sehr schöne 3-Zimmer Wohnung in Berlin-Mitte, 75qm, Balkon, renoviert, WBS erforderlich, ab sofort verfügbar",
			expected: true,
		},
		{
			name:     "numbers and wbs",
			title:    "2-Zimmer, 50qm, 800€, WBS",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterWBSString(tt.title)
			if result != tt.expected {
				t.Errorf("FilterWBSString(%q) = %v, want %v", tt.title, result, tt.expected)
			}
		})
	}
}

// TestExtractZIP tests the ZIP code extraction functionality
func TestExtractZIP(t *testing.T) {
	tests := []struct {
		name        string
		addr        string
		expectedZip string
		expectedOk  bool
	}{
		// Valid ZIP codes
		{
			name:        "standard address",
			addr:        "Musterstraße 1, 10115 Berlin",
			expectedZip: "10115",
			expectedOk:  true,
		},
		{
			name:        "zip at beginning",
			addr:        "12043 Berlin, Neuköllner Str. 123",
			expectedZip: "12043",
			expectedOk:  true,
		},
		{
			name:        "zip at end",
			addr:        "Alexanderplatz 1, Berlin 10178",
			expectedZip: "10178",
			expectedOk:  true,
		},
		{
			name:        "zip with surrounding text",
			addr:        "Apartment in 10243 Berlin-Friedrichshain",
			expectedZip: "10243",
			expectedOk:  true,
		},
		{
			name:        "multiple spaces",
			addr:        "Street Name    12345    Berlin",
			expectedZip: "12345",
			expectedOk:  true,
		},
		{
			name:        "zip with comma",
			addr:        "Straße 123, 10115, Berlin",
			expectedZip: "10115",
			expectedOk:  true,
		},
		{
			name:        "zip with parentheses",
			addr:        "Address (10115) Berlin",
			expectedZip: "10115",
			expectedOk:  true,
		},

		// Multiple ZIP codes - should return first one
		{
			name:        "multiple zip codes",
			addr:        "From 12043 to 10115 Berlin",
			expectedZip: "12043",
			expectedOk:  true,
		},
		{
			name:        "two addresses",
			addr:        "Str. 1, 10115 Berlin or Str. 2, 12043 Berlin",
			expectedZip: "10115",
			expectedOk:  true,
		},

		// Invalid ZIP codes
		{
			name:        "4-digit number",
			addr:        "Street 1234 Berlin",
			expectedZip: "",
			expectedOk:  false,
		},
		{
			name:        "6-digit number",
			addr:        "Street 123456 Berlin",
			expectedZip: "",
			expectedOk:  false,
		},
		{
			name:        "zip with letters",
			addr:        "Street 1011A Berlin",
			expectedZip: "",
			expectedOk:  false,
		},
		{
			name:        "no numbers",
			addr:        "Just a street name in Berlin",
			expectedZip: "",
			expectedOk:  false,
		},
		{
			name:        "empty string",
			addr:        "",
			expectedZip: "",
			expectedOk:  false,
		},
		{
			name:        "only whitespace",
			addr:        "   ",
			expectedZip: "",
			expectedOk:  false,
		},

		// Edge cases
		{
			name:        "zip as part of larger number",
			addr:        "Phone: 030101152345",
			expectedZip: "",
			expectedOk:  false,
		},
		{
			name:        "zip with leading zeros",
			addr:        "Address 01234 Berlin",
			expectedZip: "01234",
			expectedOk:  true,
		},
		{
			name:        "zip with dots",
			addr:        "Address 10.115 Berlin",
			expectedZip: "",
			expectedOk:  false,
		},
		{
			name:        "zip with hyphens",
			addr:        "Address 10-115 Berlin",
			expectedZip: "",
			expectedOk:  false,
		},
		{
			name:        "house number that looks like zip",
			addr:        "Musterstraße 12345, Berlin",
			expectedZip: "12345",
			expectedOk:  true, // This might be wrong - depends on context
		},
		{
			name:        "year that looks like zip",
			addr:        "Built in 12345, Musterstraße, Berlin",
			expectedZip: "12345",
			expectedOk:  true,
		},
		{
			name:        "valid berlin zip codes",
			addr:        "Kantstraße 1, 10623 Berlin",
			expectedZip: "10623",
			expectedOk:  true,
		},
		{
			name:        "valid berlin zip codes 2",
			addr:        "Friedrichstraße 1, 10117 Berlin",
			expectedZip: "10117",
			expectedOk:  true,
		},
		{
			name:        "special characters around zip",
			addr:        "Address: (10115) Berlin!",
			expectedZip: "10115",
			expectedOk:  true,
		},
		{
			name:        "zip in complex address",
			addr:        "Apartment 4A, Musterstraße 123-125, 10115 Berlin-Mitte, Germany",
			expectedZip: "10115",
			expectedOk:  true,
		},
		{
			name:        "non-german zip code",
			addr:        "123 Main St, 90210 Beverly Hills, CA",
			expectedZip: "90210",
			expectedOk:  true,
		},
		{
			name:        "umlauts in address",
			addr:        "Königstraße 1, 10115 Berlin",
			expectedZip: "10115",
			expectedOk:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zip, ok := ExtractZIP(tt.addr)
			if zip != tt.expectedZip {
				t.Errorf("ExtractZIP(%q) zip = %q, want %q", tt.addr, zip, tt.expectedZip)
			}
			if ok != tt.expectedOk {
				t.Errorf("ExtractZIP(%q) ok = %v, want %v", tt.addr, ok, tt.expectedOk)
			}
		})
	}
}

// TestExtractZIP_BerlinSpecific tests Berlin-specific ZIP code patterns
func TestExtractZIP_BerlinSpecific(t *testing.T) {
	validBerlinZips := []string{
		"10115", "10117", "10119", "10178", "10179", "10243", "10245", "10247",
		"10249", "10435", "10437", "10439", "10551", "10553", "10555", "10557",
		"10559", "10585", "10587", "10589", "10623", "10625", "10627", "10629",
		"10707", "10709", "10711", "10713", "10715", "10717", "10719", "10777",
		"10779", "10781", "10783", "10785", "10787", "10789", "10823", "10825",
		"10827", "10829", "10961", "10963", "10965", "10967", "10969", "10997",
		"10999", "12043", "12045", "12047", "12049", "12051", "12053", "12055",
		"12057", "12059", "12099", "12101", "12103", "12105", "12107", "12109",
		"12157", "12159", "12161", "12163", "12165", "12167", "12169", "12203",
		"12205", "12207", "12209", "12247", "12249", "12277", "12279", "12305",
		"12307", "12309", "12347", "12349", "12351", "12353", "12355", "12357",
		"12359", "12435", "12437", "12439", "12459", "12487", "12489", "12524",
		"12526", "12527", "12555", "12557", "12559", "12587", "12589", "12619",
		"12621", "12623", "12627", "12629", "12679", "12681", "12683", "12685",
		"12687", "12689", "13051", "13053", "13055", "13057", "13059", "13086",
		"13088", "13089", "13125", "13127", "13129", "13156", "13158", "13159",
		"13187", "13189", "13347", "13349", "13351", "13353", "13355", "13357",
		"13359", "13403", "13405", "13407", "13409", "13435", "13437", "13439",
		"13469", "13503", "13505", "13507", "13509", "13581", "13583", "13585",
		"13587", "13589", "13591", "13593", "13595", "13597", "13599", "14050",
		"14052", "14055", "14057", "14059", "14109", "14129", "14163", "14165",
		"14167", "14169", "14193", "14195", "14197", "14199",
	}

	for _, zip := range validBerlinZips {
		t.Run("berlin_zip_"+zip, func(t *testing.T) {
			addr := "Musterstraße 1, " + zip + " Berlin"
			extractedZip, ok := ExtractZIP(addr)
			if !ok {
				t.Errorf("ExtractZIP should find ZIP %s in address %q", zip, addr)
			}
			if extractedZip != zip {
				t.Errorf("ExtractZIP(%q) = %q, want %q", addr, extractedZip, zip)
			}
		})
	}
}
