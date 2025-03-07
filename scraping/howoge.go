package scraping

import (
	"apartmenthunter/config"
	"apartmenthunter/listings"
	"apartmenthunter/telegram"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type HowogeListing struct {
	ID       int     `json:"uid"`
	Title    string  `json:"title"`
	District string  `json:"district"`
	Rent     float64 `json:"rent"`
	Area     float64 `json:"area"`
	Rooms    int     `json:"rooms"`
	Wbs      string  `json:"wbs"`
	Link     string  `json:"link"`
	Notice   string  `json:"notice"`
}

// HowogeResponse Struct for API response
type HowogeResponse struct {
	Results []HowogeListing `json:"immoobjects"`
}

func CheckHowoge(state *listings.ScraperState, sendTelegram bool) {
	// 1. Define the Howoge API URL and create form data
	apiURL := config.HowogeURL
	formData := url.Values{
		"tx_howrealestate_json_list[action]": {"immoList"},
		"tx_howrealestate_json_list[page]":   {"1"},
		"tx_howrealestate_json_list[limit]":  {"50"},
		"tx_howrealestate_json_list[lang]":   {""},
		"tx_howrealestate_json_list[rooms]":  {"1"},
	}
	formData.Add("tx_howrealestate_json_list[kiez][]", "Friedrichshain-Kreuzberg")
	formData.Add("tx_howrealestate_json_list[kiez][]", "Neukölln")
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(formData.Encode()))
	if err != nil {
		log.Printf("Failed to create Howoge request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Referer", "https://www.howoge.de")
	req.Header.Set("Origin", "https://www.howoge.de")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Howoge: Failed to fetch Howoge listings: %v", err)
		return
	}
	defer resp.Body.Close()

	// 6. Parse JSON response
	var data HowogeResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		log.Printf("Howoge: Error parsing Howoge JSON response: %v", err)
		return
	}

	// 7. Process listings
	for _, listing := range data.Results {
		if !state.Exists(strconv.Itoa(listing.ID)) && listing.Wbs != "ja" {
			log.Println("new howoge post", strconv.Itoa(listing.ID))
			state.MarkAsSeen(strconv.Itoa(listing.ID))

			// Google Maps link
			encodedAddr := url.QueryEscape(listing.Title)
			mapsLink := fmt.Sprintf("https://www.google.com/maps/search/?api=1&query=%s", encodedAddr)

			// Full listing link
			listingLink := fmt.Sprintf("https://www.howoge.de%s", listing.Link)

			// --- Format Telegram message ---
			htmlMsg := fmt.Sprintf(`<b>Howoge Listing</b>

<b>Notice:</b> %s
<b>District:</b> %s
<b>Adress:</b> %s
<b>Rent:</b> %.2f €
<b>Area:</b> %.2f m²
<b>Rooms:</b> %d

<a href="%s">View in Google Maps</a>

<a href="%s">View Listing</a>`,
				listing.Notice, listing.District, listing.Title, listing.Rent, listing.Area,
				listing.Rooms, mapsLink, listingLink,
			)

			if sendTelegram {
				telegram.SendTelegramMessage(htmlMsg)
			}
		}
	}
}
