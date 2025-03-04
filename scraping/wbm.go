package scraping

import (
	"apartmenthunter/config"
	"apartmenthunter/listings"
	"apartmenthunter/telegram"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func CheckWbm(seenListings map[string]bool) {
	// 1. Fetch Wbm page
	resp, err := http.Get(config.WbmURL)
	if err != nil {
		log.Printf("Failed to wbm fetch page: %v", err)
		return
	}
	defer resp.Body.Close()

	// 2. Parse HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatalf("Error parsing HTML: %v", err)
	}

	doc.Find("div.row.openimmo-search-list-item").Each(func(i int, s *goquery.Selection) {
		postID, exists := s.Attr("data-id")
		if !exists {
			return // Skip if no data-id is found
		}
		if !seenListings[postID] { // If this listing is new
			seenListings[postID] = true // Mark as seen

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

			// 5. Send Telegram message
			telegram.SendTelegramMessage(htmlMsg)

			// 6. Append listing to file
			listings.AppendListing(config.WbmFile, postID)
		}
	})

	//noOffersText := "LEIDER HABEN WIR DERZEIT KEINE VERFÜGBAREN ANGEBOTE"
	//pageText := doc.Text()
	//
	//if !strings.Contains(strings.ToUpper(pageText), noOffersText) {
	//	telegram.SendTelegramMessage(fmt.Sprintf("WBM has a new offer: %s", config.WbmURL))
	//}
}
