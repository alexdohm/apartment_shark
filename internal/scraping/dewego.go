package scraping

import (
	"apartmenthunter/internal/bot"
	"apartmenthunter/internal/config"
	"apartmenthunter/internal/store"
	"apartmenthunter/internal/telegram"
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func CheckDewego(ctx context.Context, state *store.ScraperState, sendTelegram bool) {
	formData := url.Values{
		"tx_openimmo_immobilie[__referrer][@extension]":  {"Openimmo"},
		"tx_openimmo_immobilie[__referrer][@controller]": {"Immobilie"},
		"tx_openimmo_immobilie[__referrer][@action]":     {"search"},
		"tx_openimmo_immobilie[search]":                  {"search"},
		"tx_openimmo_immobilie[page]":                    {"1"},
		"tx_openimmo_immobilie[warmmiete_start]":         {"600"},
		"tx_openimmo_immobilie[warmmiete_end]":           {"1000"},
		"tx_openimmo_immobilie[wbsSozialwohnung]":        {"0"},
		"tx_openimmo_immobilie[distance]":                {"1"},
		"tx_openimmo_immobilie[sortBy]":                  {"immobilie_preise_warmmiete"},
		"tx_openimmo_immobilie[sortOrder]":               {"asc"},
	}

	// Add multiple values for regional filters
	formData.Add("tx_openimmo_immobilie[regionalerZusatz][]", "friedrichshain-kreuzberg")
	formData.Add("tx_openimmo_immobilie[regionalerZusatz][]", "neukolln")

	req, err := http.NewRequest("POST", config.DewegoURL, strings.NewReader(formData.Encode()))
	if err != nil {
		log.Printf("Dewego: Failed to create request: %v", err)
		return
	}
	bot.GenerateGeneralRequestHeaders(req, "https://www.degewo.de", "https://www.degewo.de/", true, false)

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

	// currently i don't have access to zip codes on this page
	doc.Find("article[id^=immobilie-list-item]").Each(func(i int, s *goquery.Selection) {
		postID, _ := s.Attr("id")
		if !state.Exists(postID) {
			log.Println("new dewego post", postID)
			state.MarkAsSeen(postID)

			address := strings.TrimSpace(s.Find("span.article__meta").Text())
			parts := strings.Split(address, "|")
			street := strings.TrimSpace(parts[0])
			neighborhood := ""
			if len(parts) > 1 {
				neighborhood = strings.TrimSpace(parts[1])
			}
			fullAddress := fmt.Sprintf("%s, %s, Berlin", street, neighborhood)

			size := strings.TrimSpace(s.Find("ul.article__properties li:nth-child(2) span.text").Text())
			size = strings.TrimSuffix(size, " m²")

			rent := strings.TrimSpace(s.Find("div.article__price-tag span.price").Text())
			rent = strings.TrimSuffix(rent, " €")

			// Extract listing link
			listingLink, exists := s.Find("a[target=_blank]").Attr("href")
			if !exists {
				listingLink = "No link available"
			} else {
				listingLink = "https://www.degewo.de" + listingLink
			}
			encodedAddr := url.QueryEscape(fullAddress)
			mapsLink := fmt.Sprintf("https://www.google.com/maps/search/?api=1&query=%s", encodedAddr)

			if sendTelegram {
				err := telegram.Send(ctx, &telegram.TelegramInfo{
					Address:     address,
					Size:        size,
					Rent:        rent,
					MapLink:     mapsLink,
					ListingLink: listingLink,
				}, "Dewego")
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	})
}
