package config

import (
	"log"
	"regexp"
	"strconv"
	"strings"
)

// IsListingWithinFilter checks to see if a listing is within a filter. if so,
// return true.
func IsListingWithinFilter(address string, size float64, rent float64) bool {
	zip, ok := extractZip(address)
	if !ok {
		log.Println("error extracting zip")
		return false
	}
	//log.Println(zip, size, rent)
	if size >= MinSqm && rent >= MinWarm && size <= MaxWarm && isZipAllowed(zip) {
		return true
	}
	return false
}

func extractZip(address string) (string, bool) {
	re := regexp.MustCompile(`\b\d{5}\b`)
	match := re.FindString(address)
	if match == "" {
		return "", false
	}
	return match, true
}

func ParseFloat(input interface{}) float64 {
	switch v := input.(type) {
	case string:
		s := strings.TrimSpace(v)

		// Detect comma as decimal separator if it's near the end (typical German style)
		if comma := strings.LastIndex(s, ","); comma != -1 && len(s)-comma <= 3 {
			s = strings.ReplaceAll(s, ".", "")  // remove thousands separator
			s = strings.ReplaceAll(s, ",", ".") // convert decimal comma to dot
		}
		parsedFloat, err := strconv.ParseFloat(s, 64)
		if err != nil {
			log.Println("error parsing float", err)
		}
		return parsedFloat

	case int:
		return float64(v)

	case float64:
		return v

	case float32:
		return float64(v)

	default:
		log.Println("error parsing float", v)
		return 0
	}
}

func isZipAllowed(zip string) bool {
	for _, allowedZip := range ZipCodes {
		if zip == allowedZip {
			return true
		}
	}
	return false
}
