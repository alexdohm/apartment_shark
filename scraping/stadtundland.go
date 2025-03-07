package scraping

import (
	"apartmenthunter/config"
	"apartmenthunter/listings"
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

func CheckStadtUndLand(state *listings.ScraperState, sendTelegram bool) {
	log.Println("starting stadt und land")
	payload := map[string]interface{}{
		"district": "Neukölln Nord",
		"offset":   0,
		"cat":      "wohnung",
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
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0")

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
		state.MarkAsSeen(listingID)

		// Construct full address
		fullAddress := fmt.Sprintf("%s %s, %s %s",
			listing.Address.Street, listing.Address.HouseNumber, listing.Address.PostalCode, listing.Address.City)

		// Construct listing link
		encodedID := url.QueryEscape(listingID)
		listingLink := fmt.Sprintf("https://stadtundland.de/wohnungssuche/%s", encodedID)

		// Construct Google Maps link
		encodedAddr := url.QueryEscape(fullAddress)
		mapsLink := fmt.Sprintf("https://www.google.com/maps/search/?api=1&query=%s", encodedAddr)

		// 6. Format Telegram Message
		htmlMsg := fmt.Sprintf(`<b>Stadt und Land Listing</b>

<b>Address:</b> %s
<b>Size:</b> %s m²
<b>Rent:</b> %s €

<a href="%s">View Map</a>

<a href="%s">View Listing</a>`,

			fullAddress, listing.Details.Area, listing.Costs.Rent, mapsLink, listingLink,
		)

		if sendTelegram {
			telegram.SendTelegramMessage(htmlMsg)
		}
	}
}
