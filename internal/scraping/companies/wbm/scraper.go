package wbm

import (
	"apartmenthunter/internal/config"
	"apartmenthunter/internal/scraping/common"
	"apartmenthunter/internal/telegram"
	"bytes"
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/url"
	"strings"
)

type scraperCtx struct {
	*common.BaseScraper
}

func Scrape(ctx context.Context, base *common.BaseScraper, sendTelegram bool) error {
	s := scraperCtx{base}

	listings, err := s.fetchListings(ctx)
	if err != nil {
		return fmt.Errorf("fetching wbm listings: %w", err)
	}

	for _, listing := range listings {
		telegramStruct := s.convertToTelegramListing(listing)

		if !s.State.Exists(listing.ID) {
			log.Printf("New WBM post: %s", listing.ID)
			s.State.MarkAsSeen(listing.ID)
			if sendTelegram {
				err := telegram.Send(ctx, telegramStruct)
				if err != nil {
					return fmt.Errorf("failed to send wbm post: %w", err)
				}
			}
		}
	}
	return nil
}

func (s *scraperCtx) fetchListings(ctx context.Context) ([]WBMListing, error) {
	headers := s.HeaderGenerator.GenerateGeneralRequestHeaders("", "", false, false)

	resp, err := s.HTTPClient.Get(ctx, config.WbmURL, headers)
	if err != nil {
		return nil, fmt.Errorf("error making get request: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP error: status code %d", resp.StatusCode)
	}

	var listings []WBMListing
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body))
	if err != nil {
		return nil, fmt.Errorf("error parsing HTML: %v", err)
	}

	doc.Find("div.row.openimmo-search-list-item").Each(func(i int, s *goquery.Selection) {
		postID, exists := s.Attr("data-id")
		if !exists {
			return // Skip if no data-id is found
		}

		address := strings.TrimSpace(s.Find("div.address").Text())
		rent := extractValue(s, "div.main-property-value.main-property-rent", " €")
		size := extractValue(s, "div.main-property-value.main-property-size", " m²")

		listingLink := fmt.Sprintf("%s#%s", config.WbmURL, postID)

		listings = append(listings, WBMListing{
			ID:      postID,
			Address: address,
			Size:    size,
			Rent:    rent,
			Link:    listingLink,
		})
	})
	return listings, nil
}

func extractValue(s *goquery.Selection, selector, suffix string) string {
	text := strings.TrimSpace(s.Find(selector).Text())
	return strings.TrimSuffix(text, suffix)
}

func (s *scraperCtx) convertToTelegramListing(listing WBMListing) *telegram.TelegramInfo {
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
