package gewobag

import (
	"apartmenthunter/internal/config"
	"apartmenthunter/internal/scraping/common"
	"apartmenthunter/internal/telegram"
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/url"
	"regexp"
	"strings"
)

type scraperCtx struct {
	*common.BaseScraper
}

func Scrape(ctx context.Context, base *common.BaseScraper, sendTelegram bool) error {
	s := scraperCtx{base}

	listings, err := s.fetchListings(ctx)
	if err != nil {
		return fmt.Errorf("fetching gewobag listings: %w", err)
	}

	for _, listing := range listings {
		telegramStruct := s.convertToTelegramListing(listing)

		if !s.State.Exists(listing.ID) {
			log.Printf("New Gewobag post: %s", listing.ID)
			s.State.MarkAsSeen(listing.ID)
			if sendTelegram {
				err := telegram.Send(ctx, telegramStruct)
				if err != nil {
					return fmt.Errorf("failed to send gewobag post: %w", err)
				}
			}
		}
	}
	return nil
}

func (s *scraperCtx) fetchListings(ctx context.Context) ([]GewobagListing, error) {
	headers := s.HeaderGenerator.GenerateGeneralRequestHeaders("", "", false, false)

	resp, err := s.HTTPClient.Get(ctx, config.GewobagURL, headers)
	if err != nil {
		return nil, fmt.Errorf("error making get request: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP error: status code %d", resp.StatusCode)
	}

	var listings []GewobagListing
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(resp.Body)))
	if err != nil {
		return nil, fmt.Errorf("error parsing HTML: %v", err)
	}

	doc.Find("article[id^=post-]").Each(func(i int, s *goquery.Selection) {
		postID, _ := s.Attr("id")

		addressText := strings.TrimSpace(s.Find("tr.angebot-address td address").Text())
		area := strings.TrimSpace(s.Find("tr.angebot-area td").Text())
		size, ok := extractSize(area)
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

		listings = append(listings, GewobagListing{
			ID:      postID,
			Address: addressText,
			Size:    size,
			Rent:    cost,
			Link:    listingLink,
		})
	})
	return listings, nil
}

func extractSize(input string) (string, bool) {
	re := regexp.MustCompile(`\d{1,3},\d{1,2}`)
	match := re.FindString(input)
	if match == "" {
		return "", false
	}
	return match, true
}

func (s *scraperCtx) convertToTelegramListing(listing GewobagListing) *telegram.TelegramInfo {
	encodedAddr := url.QueryEscape(listing.Address)
	mapsLink := fmt.Sprintf("https://www.google.com/maps/search/?api=1&query=%s", encodedAddr)

	return &telegram.TelegramInfo{
		Address:     listing.Address,
		Size:        listing.Size,
		Rent:        listing.Rent,
		MapLink:     mapsLink,
		ListingLink: listing.Link,
		Site:        s.GetName(),
	}
}
