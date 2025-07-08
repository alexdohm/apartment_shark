package common

import (
	"regexp"
	"strings"
)

type UserConfig struct {
	UserID      string
	ZipCodes    []string
	WbsRequired bool
	MinSqm      int
	MaxSqm      int
	MinPrice    int
	MaxPrice    int
}

type FilterConfig struct {
	Users []UserConfig
}

func FilterWBSString(title string) bool {
	t := strings.ToLower(title)

	// “Without WBS” phrasings
	neg := []string{
		"ohne wbs",
		"ohne wohnberechtigungsschein",
		"wbs-frei", "wbs frei",
		"o. wbs",
		"kein wbs", "keine wbs",
		"wbs nicht erforderlich",
		"ohne wohnberechtigung",
	}
	for _, n := range neg {
		if strings.Contains(t, n) {
			return false
		}
	}

	pos := []string{
		"wohnberechtigungsschein",
		"nur mit wbs", "mit wbs",
		"wbs pflicht", "wbs erforderlich",
		"wbs",
	}
	for _, p := range pos {
		if strings.Contains(t, p) {
			return true
		}
	}

	return false //if no wbs then just return false
}

// ExtractZIP finds the first German ZIP code (5 digits) in addr.
// Returns the ZIP and true on success, or "" and false if none found.
func ExtractZIP(addr string) (string, bool) {
	zipRe := regexp.MustCompile(`\b\d{5}\b`)
	if zip := zipRe.FindString(addr); zip != "" {
		return zip, true
	}
	return "", false
}
