package scraping

import (
	"apartmenthunter/internal/bot"
	"apartmenthunter/internal/config"
	"apartmenthunter/internal/store"
	"apartmenthunter/internal/telegram"
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

func CheckGewobag(ctx context.Context, state *store.ScraperState, sendTelegram bool) {
	// Create HTTP client
	client := &http.Client{}

	// Create a new GET request
	req, err := http.NewRequest("GET", config.GewobagURL, nil)
	if err != nil {
		log.Printf("Gewobag: Failed to create request: %v", err)
		return
	}
	bot.GenerateGeneralRequestHeaders(req, "", "", false, false)

	// Perform the request
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Gewobag: Failed to fetch page: %v", err)
		return
	}
	defer resp.Body.Close()

	// Parse HTML with goquery
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("Gewobag: Error parsing HTML: %v", err)
		return
	}

	// for each article with ID like "post-1234"
	doc.Find("article[id^=post-]").Each(func(i int, s *goquery.Selection) {
		postID, _ := s.Attr("id")

		if !state.Exists(postID) {
			log.Println("new Gewobag post", postID)
			state.MarkAsSeen(postID)

			addressText := strings.TrimSpace(s.Find("tr.angebot-address td address").Text())
			area := strings.TrimSpace(s.Find("tr.angebot-area td").Text())
			size, ok := ExtractSize(area)
			if !ok {
				log.Println("Error extracting size", area)
			}

			cost := strings.TrimSpace(s.Find("tr.angebot-kosten td").Text())
			cost = strings.TrimSuffix(cost, "€")
			cost = strings.TrimPrefix(cost, "ab ")

			listingLink, found := s.Find("a.read-more-link").Attr("href")
			if !found {
				listingLink = "no link found"
			}
			encodedAddr := url.QueryEscape(addressText)
			mapsLink := fmt.Sprintf("https://www.google.com/maps/search/?api=1&query=%s", encodedAddr)

			if sendTelegram && config.IsListingWithinFilter(addressText, config.ParseFloat(size), config.ParseFloat(cost)) {
				err := telegram.Send(ctx, &telegram.TelegramInfo{
					Address:     addressText,
					Size:        size,
					Rent:        cost,
					MapLink:     mapsLink,
					ListingLink: listingLink,
				}, "Gewobag")
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	})
}

func ExtractSize(input string) (string, bool) {
	re := regexp.MustCompile(`\d{1,3},\d{1,2}`)
	match := re.FindString(input)
	if match == "" {
		return "", false
	}
	return match, true
}
