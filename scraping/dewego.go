package scraping

import (
	"apartmenthunter/config"
	"apartmenthunter/listings"
	"apartmenthunter/telegram"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func CheckDewego(seenListings map[string]bool) {
	formData := url.Values{
		"tx_openimmo_immobilie[__referrer][@extension]":  {"Openimmo"},
		"tx_openimmo_immobilie[__referrer][@controller]": {"Immobilie"},
		"tx_openimmo_immobilie[__referrer][@action]":     {"search"},
		"tx_openimmo_immobilie[search]":                  {"search"},
		"tx_openimmo_immobilie[page]":                    {"1"},
		"tx_openimmo_immobilie[warmmiete_start]":         {"500"},
		"tx_openimmo_immobilie[warmmiete_end]":           {"1000"},
		"tx_openimmo_immobilie[wbsSozialwohnung]":        {"0"},
		"tx_openimmo_immobilie[distance]":                {"1"},
		"tx_openimmo_immobilie[sortBy]":                  {"immobilie_preise_warmmiete"},
		"tx_openimmo_immobilie[sortOrder]":               {"asc"},
	}

	// Add multiple values for regional filters
	formData.Add("tx_openimmo_immobilie[regionalerZusatz][]", "friedrichshain-kreuzberg")
	formData.Add("tx_openimmo_immobilie[regionalerZusatz][]", "neukolln")
	formData.Add("tx_openimmo_immobilie[regionalerZusatz][]", "tempelhof-schoneberg")

	req, err := http.NewRequest("POST", config.DewegoURL, strings.NewReader(formData.Encode()))
	if err != nil {
		log.Printf("Dewego: Failed to create request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Referer", "https://www.degewo.de/immosuche")
	req.Header.Set("Origin", "https://www.degewo.de")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Dewego: Failed to fetch Dewego listings: %v", err)
		return
	}
	defer resp.Body.Close()

	// Read the entire response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Dewego: Failed to read response: %v", err)
		return
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		log.Printf("Dewego: Error parsing HTML: %v", err)
		return
	}

	doc.Find("article[id^=immobilie-list-item]").Each(func(i int, s *goquery.Selection) {
		postID, _ := s.Attr("id") // Get the unique listing ID

		if !seenListings[postID] { // If this listing is new
			seenListings[postID] = true // Mark as seen

			// Extract listing details
			title := strings.TrimSpace(s.Find("h2.article__title").Text())
			address := strings.TrimSpace(s.Find("span.article__meta").Text())

			// Extract rooms, size, availability
			rooms := strings.TrimSpace(s.Find("ul.article__properties li:nth-child(1) span.text").Text())
			size := strings.TrimSpace(s.Find("ul.article__properties li:nth-child(2) span.text").Text())

			// Extract price
			price := strings.TrimSpace(s.Find("div.article__price-tag span.price").Text())

			// Extract listing link
			link, exists := s.Find("a[target=_blank]").Attr("href")
			if !exists {
				link = "No link available"
			} else {
				link = "https://www.degewo.de" + link // Make it a full URL
			}

			// Build Google Maps link from address
			encodedAddr := url.QueryEscape(address)
			mapsLink := fmt.Sprintf("https://www.google.com/maps/search/?api=1&query=%s", encodedAddr)

			// 4. Format Telegram message
			htmlMsg := fmt.Sprintf(`<b>New Dewego Listing</b>

<b>Title:</b> %s
<b>Address:</b> %s
<b>Rooms:</b> %s
<b>Size:</b> %s
<b>Rent:</b> %s

<a href="%s">View in Google Maps</a>

<a href="%s">View Listing</a>`,
				title, address, rooms, size, price,
				mapsLink, link,
			)

			// 5. Send Telegram notification
			//log.Println(htmlMsg) //todo get the request right for location here
			telegram.SendTelegramMessage(htmlMsg)

			//6. Append listing to file
			listings.AppendListing(config.DewegoFile, postID)
		}
	})

}
