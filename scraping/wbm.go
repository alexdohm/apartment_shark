package scraping

import (
	"apartmenthunter/bot"
	"apartmenthunter/config"
	"apartmenthunter/store"
	"apartmenthunter/telegram"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func CheckWbm(state *store.ScraperState, sendTelegram bool) {
	// Create HTTP client
	client := &http.Client{}

	// Create a new GET request
	req, err := http.NewRequest("GET", config.WbmURL, nil)
	if err != nil {
		log.Printf("WBM: Failed to create request: %v", err)
		return
	}
	bot.GenerateGeneralRequestHeaders(req, "", "", false, false)

	// Perform the request
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("WBM: Failed to fetch page: %v", err)
		return
	}
	defer resp.Body.Close()

	// Parse HTML with goquery
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("WBM: Error parsing HTML: %v", err)
		return
	}

	doc.Find("div.row.openimmo-search-list-item").Each(func(i int, s *goquery.Selection) {
		postID, exists := s.Attr("data-id")
		if !exists {
			return // Skip if no data-id is found
		}
		if !state.Exists(postID) { // If this listing is new
			log.Println("new WBM post", postID)
			state.MarkAsSeen(postID)

			// Extract details
			region := strings.TrimSpace(s.Find("div.area").Text())
			address := strings.TrimSpace(s.Find("div.address").Text())
			rent := strings.TrimSpace(s.Find("div.main-property-value.main-property-rent").Text())
			size := strings.TrimSpace(s.Find("div.main-property-value.main-property-size").Text())
			rooms := strings.TrimSpace(s.Find("div.main-property-value.main-property-rooms").Text())

			// Google Maps link for address search
			encodedAddr := url.QueryEscape(address)
			mapsLink := fmt.Sprintf("https://www.google.com/maps/search/?api=1&query=%s", encodedAddr)

			// WBM listing link (assuming individual listings have unique links)
			listingLink := fmt.Sprintf("%s#%s", config.WbmURL, postID)

			// Format the Telegram message
			htmlMsg := fmt.Sprintf(`<b>WBM Listing</b>

<b>Region:</b> %s
<b>Adresse:</b> %s
<b>Miete:</b> %s
<b>Größe:</b> %s
<b>Zimmer:</b> %s

<a href="%s">View in Google Maps</a>

<a href="%s">View Listing</a>`,
				region, address, rent, size, rooms,
				mapsLink, listingLink,
			)

			if sendTelegram {
				telegram.SendTelegramMessage(htmlMsg)
			}
		}
	})
}
