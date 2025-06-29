package howoge

import (
	"apartmenthunter/internal/config"
	"apartmenthunter/internal/scraping/common"
	"context"
	"encoding/json"
	"fmt"
)

type scraperCtx struct {
	*common.BaseScraper
}

func Scrape(ctx context.Context, base *common.BaseScraper) ([]common.Listing, error) {
	s := scraperCtx{base}

	listings, err := s.fetchListings(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetching howoge listings: %w", err)
	}

	return listings, nil
}

func (s *scraperCtx) fetchListings(ctx context.Context) ([]common.Listing, error) {
	formData := s.buildFormData()
	headers := s.HeaderGenerator.GenerateGeneralRequestHeaders("https://www.howoge.de", "https://www.howoge.de", true, false)

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

	var listings []common.Listing
	for _, listing := range data.Results {
		listings = append(listings, common.Listing{
			ID:      fmt.Sprintf("%d", listing.ID),
			Company: "Howoge",
			Price:   fmt.Sprintf("%.2f", listing.Rent),
			Size:    fmt.Sprintf("%.2f", listing.Size),
			Address: listing.Address,
			URL:     fmt.Sprintf("https://www.howoge.de%s", listing.Link),
		})
	}
	return listings, nil
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
