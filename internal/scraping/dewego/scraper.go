package dewego

import (
	"apartmenthunter/internal/config"
	"apartmenthunter/internal/scraping/common"
	"apartmenthunter/internal/telegram"
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/url"
	"strings"
)

type scraperCtx struct {
	*common.BaseScraper
}

func Scrape(ctx context.Context, base *common.BaseScraper, sendTelegram bool) error {
	s := scraperCtx{base}

	listings, err := s.fetchListings(ctx)
	if err != nil {
		return fmt.Errorf("fetching dewego listings: %w", err)
	}

	for _, listing := range listings {
		telegramStruct := s.convertToTelegramListing(listing)

		if !s.State.Exists(listing.ID) {
			log.Printf("New Dewego post: %s", listing.ID)
			s.State.MarkAsSeen(listing.ID)
			if sendTelegram {
				err := telegram.Send(ctx, telegramStruct)
				if err != nil {
					return fmt.Errorf("failed to send howoge post: %w", err)
				}
			}
		}
	}
	return nil
}

func (s *scraperCtx) fetchListings(ctx context.Context) ([]DewegoListing, error) {
	formData := s.buildFormData()
	headers := s.HeaderGenerator.GenerateGeneralRequestHeaders("", "", true, false)

	resp, err := s.HTTPClient.Post(ctx, config.DewegoURL, formData, headers)
	if err != nil {
		return nil, fmt.Errorf("error making post request: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP error: status code %d", resp.StatusCode)
	}

	var listings []DewegoListing

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(resp.Body)))
	if err != nil {
		return nil, fmt.Errorf("error parsing HTML: %v", err)
	}

	doc.Find("article[id^=immobilie-list-item]").Each(func(i int, s *goquery.Selection) {
		postID, _ := s.Attr("id")

		address := strings.TrimSpace(s.Find("span.article__meta").Text())
		parts := strings.Split(address, "|")
		street := strings.TrimSpace(parts[0])
		neighborhood := ""
		if len(parts) > 1 {
			neighborhood = strings.TrimSpace(parts[1])
		}
		fullAddress := fmt.Sprintf("%s, %s, Berlin", street, neighborhood)

		size := strings.TrimSpace(s.Find("ul.article__properties li:nth-child(2) span.text").Text())
		size = strings.TrimSuffix(size, " m²")

		rent := strings.TrimSpace(s.Find("div.article__price-tag span.price").Text())
		rent = strings.TrimSuffix(rent, " €")

		listingLink, exists := s.Find("a[target=_blank]").Attr("href")
		if !exists {
			listingLink = ""
			fmt.Printf("no dewego listing link found for %s", postID)
		}

		listings = append(listings, DewegoListing{
			ID:          postID,
			Address:     fullAddress,
			Size:        size,
			Rent:        rent,
			ListingLink: listingLink,
		})
	})
	return listings, nil
}

func (s *scraperCtx) buildFormData() map[string][]string {
	formData := map[string][]string{
		"tx_openimmo_immobilie[__referrer][@extension]":  {"Openimmo"},
		"tx_openimmo_immobilie[__referrer][@controller]": {"Immobilie"},
		"tx_openimmo_immobilie[__referrer][@action]":     {"search"},
		"tx_openimmo_immobilie[search]":                  {"search"},
		"tx_openimmo_immobilie[page]":                    {"1"},
		"tx_openimmo_immobilie[warmmiete_start]":         {"600"},
		"tx_openimmo_immobilie[warmmiete_end]":           {"1000"},
		"tx_openimmo_immobilie[wbsSozialwohnung]":        {"0"},
		"tx_openimmo_immobilie[distance]":                {"1"},
		"tx_openimmo_immobilie[sortBy]":                  {"immobilie_preise_warmmiete"},
		"tx_openimmo_immobilie[sortOrder]":               {"asc"},
		"tx_openimmo_immobilie[regionalerZusatz][]": {
			"Friedrichshain-Kreuzberg",
			"neukolln",
		},
	}
	return formData
}

func (s *scraperCtx) convertToTelegramListing(listing DewegoListing) *telegram.TelegramInfo {
	encodedAddr := url.QueryEscape(listing.Address)
	mapsLink := fmt.Sprintf("https://www.google.com/maps/search/?api=1&query=%s", encodedAddr)
	listingLink := fmt.Sprintf("https://www.degewo.de%s", listing.ListingLink)

	return &telegram.TelegramInfo{
		Address:     listing.Address,
		Size:        listing.Size,
		Rent:        listing.Rent,
		MapLink:     mapsLink,
		ListingLink: listingLink,
		Site:        s.GetName(),
	}
}
