package wbm

import (
	"apartmenthunter/internal/config"
	"apartmenthunter/internal/scraping/common"
	"bytes"
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"strings"
)

func FetchListings(ctx context.Context, base *common.BaseScraper) ([]common.Listing, error) {
	headers := base.HeaderGenerator.GenerateGeneralRequestHeaders("", "", false, false)

	resp, err := base.HTTPClient.Get(ctx, config.WbmURL, headers)
	if err != nil {
		return nil, fmt.Errorf("error making get request: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP error: status code %d", resp.StatusCode)
	}

	var listings []common.Listing
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
		cost := extractValue(s, "div.main-property-value.main-property-rent", " €")
		size := extractValue(s, "div.main-property-value.main-property-size", " m²")

		relLink, exists := s.Find("div.btn-holder a").Attr("href")
		if !exists {
			log.Println("No WBM listing link found for ", postID)
		}
		listingLink := fmt.Sprintf("%s%s", "https://www.wbm.de", relLink)

		listings = append(listings, common.Listing{
			ID:      postID,
			Company: "WBM",
			Price:   cost,
			Size:    size,
			Address: address,
			URL:     listingLink,
		})
	})
	return listings, nil
}

func extractValue(s *goquery.Selection, selector, suffix string) string {
	text := strings.TrimSpace(s.Find(selector).Text())
	return strings.TrimSuffix(text, suffix)
}
