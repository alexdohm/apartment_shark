package howoge

import (
	"apartmenthunter/internal/config"
	"apartmenthunter/internal/scraping/common"
	"apartmenthunter/internal/telegram"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strconv"
)

type scraperCtx struct {
	*common.BaseScraper
}

func Scrape(ctx context.Context, base *common.BaseScraper, sendTelegram bool) error {
	s := scraperCtx{base}

	listings, err := s.fetchListings(ctx)
	if err != nil {
		return fmt.Errorf("fetching howoge listings: %w", err)
	}

	for _, listing := range listings {
		telegramStruct := s.convertToTelegramListing(listing)

		if !s.State.Exists(strconv.Itoa(listing.ID)) && listing.Wbs != "ja" {
			log.Printf("New Howoge post: %d", listing.ID)
			s.State.MarkAsSeen(strconv.Itoa(listing.ID))
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

func (s *scraperCtx) fetchListings(ctx context.Context) ([]HowogeListing, error) {
	formData := s.buildFormData()
	headers := s.HeaderGenerator.GenerateGeneralRequestHeaders("", "", true, false)

	resp, err := s.HTTPClient.Post(ctx, config.HowogeURL, formData, headers)
	if err != nil {
		return nil, fmt.Errorf("error making post request: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP error: status code %d", resp.StatusCode)
	}

	var data HowogeResponse
	if err := json.Unmarshal(resp.Body, &data); err != nil {
		return nil, fmt.Errorf("error parsing json response: %w", err)
	}
	return data.Results, nil
}

func (s *scraperCtx) buildFormData() map[string][]string {
	formData := map[string][]string{
		"tx_howrealestate_json_list[action]": {"immoList"},
		"tx_howrealestate_json_list[page]":   {"1"},
		"tx_howrealestate_json_list[limit]":  {"50"},
		"tx_howrealestate_json_list[lang]":   {""},
		"tx_howrealestate_json_list[kiez][]": {
			"Friedrichshain-Kreuzberg",
			"Neukölln",
			"Tempelhof-Schöneberg",
			"Treptow-Köpenick",
		},
	}
	return formData
}

func (s *scraperCtx) convertToTelegramListing(listing HowogeListing) *telegram.TelegramInfo {
	encodedAddr := url.QueryEscape(listing.Address)
	mapsLink := fmt.Sprintf("https://www.google.com/maps/search/?api=1&query=%s", encodedAddr)
	listingLink := fmt.Sprintf("https://www.howoge.de%s", listing.Link)

	return &telegram.TelegramInfo{
		Address:     listing.Address,
		Size:        fmt.Sprintf("%.2f", listing.Size),
		Rent:        fmt.Sprintf("%.2f", listing.Rent),
		MapLink:     mapsLink,
		ListingLink: listingLink,
		Site:        s.GetName(),
	}
}
