package scraping

import (
	"apartmenthunter/bot"
	"apartmenthunter/config"
	"apartmenthunter/store"
	"apartmenthunter/telegram"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
)

type Address struct {
	Street      string `json:"street"`
	HouseNumber string `json:"house_number"`
	PostalCode  string `json:"postal_code"`
	City        string `json:"city"`
}

type Details struct {
	Id   string `json:"immoNumber"`
	Area string `json:"livingSpace"`
}

type Costs struct {
	Rent string `json:"warmRent"`
}

type StadtUndLandListing struct {
	Title   string  `json:"headline"`
	Address Address `json:"address"`
	Details Details `json:"details"`
	Costs   Costs   `json:"costs"`
	Link    string  `json:"url"`
}

type StadtUndLandResponse struct {
	Listings []StadtUndLandListing `json:"data"`
}

func CheckStadtUndLand(state *store.ScraperState, sendTelegram bool) {
	payload := map[string]interface{}{
		"offset": 0,
		"cat":    "wohnung",
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Stadt Und Land: Error encoding JSON: %v", err)
		return
	}
	req, err := http.NewRequest("POST", config.StadtUndLandURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Stadt Und Land: Failed to create request: %v", err)
		return
	}
	bot.GenerateGeneralRequestHeaders(req, "", "", false, true)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Stadt Und Land: Failed to fetch listings: %v", err)
		return
	}
	defer resp.Body.Close()

	var data StadtUndLandResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		log.Printf("Stadt Und Land: Error parsing JSON response: %v", err)
		return
	}

	for _, listing := range data.Listings {
		listingID := listing.Details.Id

		if state.Exists(listingID) {
			continue
		}
		log.Println("new Stadt Und Land post", listingID)
		state.MarkAsSeen(listingID)

		fullAddress := fmt.Sprintf("%s %s, %s %s",
			listing.Address.Street, listing.Address.HouseNumber, listing.Address.PostalCode, listing.Address.City)

		encodedID := url.QueryEscape(listingID)
		listingLink := fmt.Sprintf("https://stadtundland.de/wohnungssuche/%s", encodedID)

		encodedAddr := url.QueryEscape(fullAddress)
		mapsLink := fmt.Sprintf("https://www.google.com/maps/search/?api=1&query=%s", encodedAddr)

		if sendTelegram {
			telegram.GenerateTelegramMessage(&telegram.TelegramInfo{
				Address:     fullAddress,
				Size:        listing.Details.Area,
				Rent:        listing.Costs.Rent,
				MapLink:     mapsLink,
				ListingLink: listingLink,
			}, "Stadt Und Land")
		}
	}
}
