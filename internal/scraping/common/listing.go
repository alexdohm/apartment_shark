package common

import (
	"apartmenthunter/internal/telegram"
	"apartmenthunter/internal/users"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
)

type Listing struct {
	ID          string
	Company     string
	Price       string
	Size        string
	Address     string
	URL         string
	ZipCode     string
	WbsRequired bool
}

// ToTelegramInfo converts a listing to telegram struct
func (l Listing) ToTelegramInfo() *telegram.TelegramInfo {
	encodedAddr := url.QueryEscape(l.Address)
	mapsLink := fmt.Sprintf("https://www.google.com/maps/search/?api=1&query=%s", encodedAddr)

	return &telegram.TelegramInfo{
		Address:     l.Address,
		Size:        l.Size,
		Rent:        l.Price,
		MapLink:     mapsLink,
		ListingLink: l.URL,
		Site:        l.Company,
	}
}

func (l Listing) MatchUserConfig(userConfig *users.UserConfig) bool {
	// Check zip code, WBS, price, size
	return l.matchesZipCode(userConfig.ZipCodes) &&
		l.matchesWbsRequirement(userConfig.WbsRequired) &&
		l.matchesPriceRange(userConfig.MinPrice, userConfig.MaxPrice) &&
		l.matchesSizeRange(userConfig.MinSqm, userConfig.MaxSqm)
}

func (l Listing) matchesZipCode(allowedZipCodes []string) bool {
	if len(allowedZipCodes) == 0 {
		return true // No restriction
	}

	for _, zipCode := range allowedZipCodes {
		if l.ZipCode == zipCode {
			return true
		}
	}
	return false
}

func (l Listing) matchesWbsRequirement(userWbsRequired bool) bool {
	// If user requires WBS but listing doesn't have it, reject
	if userWbsRequired && !l.WbsRequired {
		return false
	}
	return true
}

func (l Listing) matchesPriceRange(minPrice, maxPrice int) bool {
	if minPrice == 0 && maxPrice == 0 {
		return true // No price restriction
	}

	price, err := l.parseIntFromString(l.Price)
	if err != nil {
		return false
	}

	if minPrice > 0 && price < minPrice {
		return false
	}

	if maxPrice > 0 && price > maxPrice {
		return false
	}

	return true
}

func (l Listing) matchesSizeRange(minSqm, maxSqm int) bool {
	if minSqm == 0 && maxSqm == 0 {
		return true // No size restriction
	}

	size, err := l.parseIntFromString(l.Size)
	if err != nil {
		return false
	}

	if minSqm > 0 && size < minSqm {
		return false
	}

	if maxSqm > 0 && size > maxSqm {
		return false
	}

	return true
}

func (l Listing) parseIntFromString(s string) (int, error) {
	re := regexp.MustCompile(`(\d+(?:\.\d+)?)`)
	matches := re.FindStringSubmatch(s)
	if len(matches) < 2 {
		return 0, strconv.ErrSyntax
	}

	numFloat, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, err
	}

	return int(numFloat), nil
}
