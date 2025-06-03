package scraping

import (
	"apartmenthunter/internal/bot"
	"apartmenthunter/internal/config"
	"apartmenthunter/internal/store"
	"apartmenthunter/internal/telegram"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type HowogeListing struct {
	ID      int     `json:"uid"`
	Address string  `json:"title"`
	Rent    float64 `json:"rent"`
	Size    float64 `json:"area"`
	Wbs     string  `json:"wbs"`
	Link    string  `json:"link"`
	Notice  string  `json:"notice"`
}

// HowogeResponse Struct for API response
type HowogeResponse struct {
	Results []HowogeListing `json:"immoobjects"`
}

func CheckHowoge(ctx context.Context, state *store.ScraperState, sendTelegram bool) {
	apiURL := config.HowogeURL
	formData := url.Values{
		"tx_howrealestate_json_list[action]": {"immoList"},
		"tx_howrealestate_json_list[page]":   {"1"},
		"tx_howrealestate_json_list[limit]":  {"50"},
		"tx_howrealestate_json_list[lang]":   {""},
	}
	formData.Add("tx_howrealestate_json_list[kiez][]", "Friedrichshain-Kreuzberg")
	formData.Add("tx_howrealestate_json_list[kiez][]", "Neukölln")
	formData.Add("tx_howrealestate_json_list[kiez][]", "Tempelhof-Schöneberg")
	formData.Add("tx_howrealestate_json_list[kiez][]", "Treptow-Köpenick")

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(formData.Encode()))
	if err != nil {
		log.Printf("Failed to create Howoge request: %v", err)
	}
	bot.GenerateGeneralRequestHeaders(req, "", "", true, false)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Howoge: Failed to fetch Howoge listings: %v", err)
		return
	}
	defer resp.Body.Close()

	// Parse JSON response
	var data HowogeResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		log.Printf("Howoge: Error parsing Howoge JSON response: %v", err)
		return
	}

	// Process listings
	for _, listing := range data.Results {
		if !state.Exists(strconv.Itoa(listing.ID)) && listing.Wbs != "ja" {
			log.Println("new howoge post", strconv.Itoa(listing.ID))

			state.MarkAsSeen(strconv.Itoa(listing.ID))
			encodedAddr := url.QueryEscape(listing.Address)
			mapsLink := fmt.Sprintf("https://www.google.com/maps/search/?api=1&query=%s", encodedAddr)
			listingLink := fmt.Sprintf("https://www.howoge.de%s", listing.Link)

			if sendTelegram && config.IsListingWithinFilter(listing.Address, config.ParseFloat(listing.Size), config.ParseFloat(listing.Rent)) {

				err := telegram.Send(ctx, &telegram.TelegramInfo{
					Address:     listing.Address,
					Size:        fmt.Sprintf("%.2f", listing.Size),
					Rent:        fmt.Sprintf("%.2f", listing.Rent),
					MapLink:     mapsLink,
					ListingLink: listingLink,
				}, "Howoge")
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}
}
