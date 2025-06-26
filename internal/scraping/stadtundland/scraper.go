package stadtundland

import (
	"apartmenthunter/internal/config"
	"apartmenthunter/internal/scraping/common"
	"apartmenthunter/internal/telegram"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
)

type scraperCtx struct {
	*common.BaseScraper
}

func Scrape(ctx context.Context, base *common.BaseScraper, sendTelegram bool) error {
	s := scraperCtx{base}

	listings, err := s.fetchListings(ctx)
	if err != nil {
		return err
	}

	for _, listing := range listings {
		telegramStruct := s.convertToTelegramListing(listing)
		id := listing.Details.Id
		if !s.State.Exists(id) {
			log.Printf("New Stadt Und Land post: %s", id)
			s.State.MarkAsSeen(id)
			if sendTelegram {
				err := telegram.Send(ctx, telegramStruct)
				if err != nil {
					return fmt.Errorf("failed to send stadt und land post: %w", err)
				}
			}
		}
	}
	return nil
}

func (s *scraperCtx) fetchListings(ctx context.Context) ([]StadtUndLandListing, error) {
	formData := s.buildFormData()
	headers := s.HeaderGenerator.GenerateGeneralRequestHeaders("", "", false, true)

	resp, err := s.HTTPClient.PostJSON(ctx, config.StadtUndLandURL, formData, headers)
	if err != nil {
		return nil, fmt.Errorf("error making post request: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP error: status code %d", resp.StatusCode)
	}

	var data StadtUndLandResponse
	if err := json.Unmarshal(resp.Body, &data); err != nil {
		return nil, fmt.Errorf("stadt und land: error parsing json response: %w", err)
	}
	return data.Listings, nil
}

func (s *scraperCtx) buildFormData() []byte {
	formData := map[string]interface{}{
		"offset": 0,
		"cat":    "wohnung",
	}
	jsonData, err := json.Marshal(formData)
	if err != nil {
		log.Printf("Stadt Und Land: Error encoding JSON: %v", err)
		return nil
	}
	return jsonData
}

func (s *scraperCtx) convertToTelegramListing(listing StadtUndLandListing) *telegram.TelegramInfo {
	listingID := listing.Details.Id

	fullAddress := fmt.Sprintf("%s %s, %s %s",
		listing.Address.Street, listing.Address.HouseNumber, listing.Address.PostalCode, listing.Address.City)

	encodedID := url.QueryEscape(listingID)
	listingLink := fmt.Sprintf("https://stadtundland.de/wohnungssuche/%s", encodedID)

	encodedAddr := url.QueryEscape(fullAddress)
	mapsLink := fmt.Sprintf("https://www.google.com/maps/search/?api=1&query=%s", encodedAddr)

	return &telegram.TelegramInfo{
		Address:     fullAddress,
		Size:        listing.Details.Area,
		Rent:        listing.Costs.Rent,
		MapLink:     mapsLink,
		ListingLink: listingLink,
		Site:        s.GetName(),
	}
}
