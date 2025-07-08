package stadtundland

import (
	"apartmenthunter/internal/config"
	"apartmenthunter/internal/scraping/common"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
)

func FetchListings(ctx context.Context, base *common.BaseScraper) ([]common.Listing, error) {
	formData := buildFormData()
	headers := base.HeaderGenerator.GenerateGeneralRequestHeaders("", "", false, true)

	resp, err := base.HTTPClient.PostJSON(ctx, config.StadtUndLandURL, formData, headers)
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

	var listings []common.Listing
	for _, listing := range data.Listings {
		listings = append(listings, common.Listing{
			ID:      listing.Details.Id,
			Company: "Stadt Und Land",
			Price:   listing.Costs.Rent,
			Size:    listing.Details.Area,
			Address: fmt.Sprintf("%s %s, %s %s",
				listing.Address.Street, listing.Address.HouseNumber, listing.Address.PostalCode, listing.Address.City),
			URL:         fmt.Sprintf("https://stadtundland.de/wohnungssuche/%s", url.QueryEscape(listing.Details.Id)),
			ZipCode:     listing.Address.PostalCode,
			WbsRequired: common.FilterWBSString(listing.Title),
		})
	}
	return listings, nil
}

func buildFormData() []byte {
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
