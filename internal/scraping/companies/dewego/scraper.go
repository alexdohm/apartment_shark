package dewego

import (
	"apartmenthunter/internal/config"
	"apartmenthunter/internal/scraping/common"
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const pageSize = 10

func FetchListings(ctx context.Context, base *common.BaseScraper) ([]common.Listing, error) {
	var (
		allListings []common.Listing
		offset      int
		totalItems  int
	)

	for {
		doc, err := fetchSearchPage(ctx, base, offset, pageSize)
		if err != nil {
			return allListings, err
		}

		if offset == 0 {
			totalItems = parseTotalListings(doc)
			if totalItems == 0 {
				break // No results found
			}
		}

		listingsOnPage := parseListings(doc)
		allListings = append(allListings, listingsOnPage...)

		offset += pageSize

		if offset >= totalItems {
			break
		}

		select {
		case <-ctx.Done():
			return allListings, ctx.Err()
		case <-time.After(250 * time.Millisecond): // reduce load
		}
	}
	return allListings, nil
}

// fetchSearchPage sends the POST request and returns the parsed goquery document.
func fetchSearchPage(ctx context.Context, base *common.BaseScraper, offset, limit int) (*goquery.Document, error) {
	formData := buildFormData()
	formData["tx_openimmo_immobilie[page]"] = []string{strconv.Itoa((offset / limit) + 1)}

	headers := base.HeaderGenerator.GenerateGeneralRequestHeaders("", "", true, false)

	resp, err := base.HTTPClient.Post(ctx, config.DewegoURL, formData, headers)
	if err != nil {
		return nil, fmt.Errorf("POST request failed: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP status %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(resp.Body)))
	if err != nil {
		return nil, fmt.Errorf("parsing HTML failed: %w", err)
	}

	return doc, nil
}

// parseTotalListings extracts the total number of listings from the results text.
func parseTotalListings(doc *goquery.Document) int {
	text := doc.Find("#openimmo-search-result").Text()
	re := regexp.MustCompile(`(\d+)\s+Treffer`)
	matches := re.FindStringSubmatch(text)

	if len(matches) < 2 {
		return 0
	}

	total, _ := strconv.Atoi(matches[1])
	return total
}

// parseListings extracts listing information from the document.
func parseListings(doc *goquery.Document) []common.Listing {
	var listings []common.Listing

	doc.Find("article[id^=immobilie-list-item]").Each(func(_ int, s *goquery.Selection) {
		postID, _ := s.Attr("id")

		addressText := strings.TrimSpace(s.Find("span.article__meta").Text())
		addressParts := strings.Split(addressText, "|")

		street := strings.TrimSpace(addressParts[0])
		neighborhood := ""
		if len(addressParts) > 1 {
			neighborhood = strings.TrimSpace(addressParts[1])
		}

		fullAddress := fmt.Sprintf("%s, %s, Berlin", street, neighborhood)

		sizeText := strings.TrimSpace(s.Find("ul.article__properties li:nth-child(2) span.text").Text())
		size := strings.TrimSuffix(sizeText, " m²")

		rentText := strings.TrimSpace(s.Find("div.article__price-tag span.price").Text())
		rent := strings.TrimSuffix(rentText, " €")

		listingLink, exists := s.Find("a[target=_blank]").Attr("href")
		if !exists {
			listingLink = ""
		} else {
			listingLink = fmt.Sprintf("https://www.degewo.de%s", listingLink)
		}

		listings = append(listings, common.Listing{
			ID:      postID,
			Company: "Dewego",
			Price:   rent,
			Size:    size,
			Address: fullAddress,
			URL:     listingLink,
		})
	})

	return listings
}

func buildFormData() map[string][]string {
	formData := map[string][]string{
		"tx_openimmo_immobilie[__referrer][@extension]":  {"Openimmo"},
		"tx_openimmo_immobilie[__referrer][@controller]": {"Immobilie"},
		"tx_openimmo_immobilie[__referrer][@action]":     {"search"},
		"tx_openimmo_immobilie[search]":                  {"search"},
		"tx_openimmo_immobilie[page]":                    {"1"},
		//"tx_openimmo_immobilie[sortBy]":    {"immobilie_preise_warmmiete"},
		//"tx_openimmo_immobilie[sortOrder]": {"asc"},
		//"tx_openimmo_immobilie[warmmiete_start]":         {"600"},
		//"tx_openimmo_immobilie[warmmiete_end]":           {"1000"},
		//"tx_openimmo_immobilie[wbsSozialwohnung]":        {"0"},
		//"tx_openimmo_immobilie[distance]":                {"1"},
		//"tx_openimmo_immobilie[regionalerZusatz][]": {
		//	"Friedrichshain-Kreuzberg",
		//	"neukolln",
		//},
	}
	return formData
}
