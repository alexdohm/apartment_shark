package common

import (
	"apartmenthunter/internal/telegram"
	"fmt"
	"net/url"
)

type Listing struct {
	ID      string
	Company string
	Price   string
	Size    string
	Address string
	URL     string
}

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
