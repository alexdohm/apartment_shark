package telegram

import (
	"fmt"
)

type TelegramInfo struct {
	Address, Size, Rent, MapLink, ListingLink, Site string
}

func BuildHTML(info *TelegramInfo) string {
	if info == nil {
		return "Data not provided"
	}

	address := info.Address
	if address == "" {
		address = "-"
	}

	size := info.Size
	if size == "" {
		size = "-"
	}

	rent := info.Rent
	if rent == "" {
		rent = "-"
	}

	site := info.Site

	mapLink := info.MapLink
	if mapLink == "" {
		mapLink = "#"
	}

	listingLink := info.ListingLink
	if listingLink == "" {
		listingLink = "#"
	}

	return fmt.Sprintf(`<b>%s Listing</b>

<b>Address:</b> %s
<b>Size:</b> %s m²
<b>Rent:</b> %s €

<a href="%s">View Map</a>
<a href="%s">View Listing</a>`,
		site, address, size, rent, mapLink, listingLink,
	)
}
