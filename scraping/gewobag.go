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

func CheckGewobag(seenListings map[string]bool) {
	// 1. Fetch Gewobag page
	resp, err := http.Get(config.GewobagURL)
	if err != nil {
		log.Printf("Gewobag: Failed to fetch page: %v", err)
		return
	}
	defer resp.Body.Close()

	// 2. Parse HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("Gewobag: Error parsing HTML: %v", err)
		return
	}

	// for each article with ID like "post-1234"
	doc.Find("article[id^=post-]").Each(func(i int, s *goquery.Selection) {
		postID, _ := s.Attr("id")

		if !seenListings[postID] {
			seenListings[postID] = true //add postID to the map

			// --- Extract listing details ---
			region := strings.TrimSpace(s.Find("tr.angebot-region td").Text())
			addressText := strings.TrimSpace(s.Find("tr.angebot-address td address").Text())
			titleText := strings.TrimSpace(s.Find("tr.angebot-address td h3.angebot-title").Text())
			area := strings.TrimSpace(s.Find("tr.angebot-area td").Text())
			availability := strings.TrimSpace(s.Find("tr.availability td").Text())
			cost := strings.TrimSpace(s.Find("tr.angebot-kosten td").Text())

			// The "read more" link
			readMoreLink, found := s.Find("a.read-more-link").Attr("href")
			if !found {
				readMoreLink = "no link found"
			}

			// Build Google Maps link from address
			encodedAddr := url.QueryEscape(addressText)
			mapsLink := fmt.Sprintf("https://www.google.com/maps/search/?api=1&query=%s", encodedAddr)

			// --- Build HTML message ---
			htmlMsg := fmt.Sprintf(`<b>Gewobag</b>

<b>Region:</b> %s
<b>Titel:</b> %s
<b>Größe:</b> %s
<b>Verfügbarkeit:</b> %s
<b>Kosten:</b> %s

<b>Adresse:</b> %s
<a href="%s">View in google maps</a>

<a href="%s">Link to apply</a>`,
				region, titleText, area, availability, cost,
				addressText, mapsLink,
				readMoreLink,
			)

			// 3. send message
			telegram.SendTelegramMessage(htmlMsg)

			// 4. Append listing to file
			listings.AppendListing(config.GewobagFile, postID)
		}
	})
}
