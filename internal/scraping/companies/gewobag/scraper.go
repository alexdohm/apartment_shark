package gewobag

import (
	"apartmenthunter/internal/config"
	"apartmenthunter/internal/scraping/common"
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"regexp"
	"strings"
)

func FetchListings(ctx context.Context, base *common.BaseScraper) ([]common.Listing, error) {
	headers := base.HeaderGenerator.GenerateGeneralRequestHeaders("", "", false, false)

	resp, err := base.HTTPClient.Get(ctx, config.GewobagURL, headers)
	if err != nil {
		return nil, fmt.Errorf("error making get request: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP error: status code %d", resp.StatusCode)
	}

	var listings []common.Listing
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(resp.Body)))
	if err != nil {
		return nil, fmt.Errorf("error parsing HTML: %v", err)
	}

	doc.Find("article[id^=post-]").Each(func(i int, s *goquery.Selection) {
		postID, _ := s.Attr("id")

		title := strings.TrimSpace(s.Find("tr.angebot-address td h3.angebot-title").Text())
		isWbs := common.FilterWBSString(title)
		addressText := strings.TrimSpace(s.Find("tr.angebot-address td address").Text())
		zip, ok := common.ExtractZIP(addressText)
		if !ok {
			log.Println("Error extracting zip", addressText)
		}
		area := strings.TrimSpace(s.Find("tr.angebot-area td").Text())
		size, ok := extractSize(area)
		if !ok {
			log.Println("[Gewobag] Error extracting size", area)
		}

		cost := strings.TrimSpace(s.Find("tr.angebot-kosten td").Text())
		cost = strings.TrimSuffix(cost, "€")
		cost = strings.TrimPrefix(cost, "ab ")

		listingLink, found := s.Find("a.read-more-link").Attr("href")
		if !found {
			listingLink = "no link found"
		}

		listings = append(listings, common.Listing{
			ID:          postID,
			Company:     "Gewobag",
			Price:       cost,
			Size:        size,
			Address:     addressText,
			URL:         listingLink,
			WbsRequired: isWbs,
			ZipCode:     zip,
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
