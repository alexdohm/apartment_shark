package main

import (
	"apartmenthunter/internal/http"
	"apartmenthunter/internal/scraping/common"
	"apartmenthunter/internal/scraping/howoge"
	"apartmenthunter/internal/store"
)

type ScraperFactory interface {
	CreateScraper(scraperType string, state *store.ScraperState) common.Scraper
}

// DefaultScraperFactory creates scrapers with shared dependencies
type DefaultScraperFactory struct {
	httpClient http.HTTPClient
}

func NewScraperFactory(httpClient http.HTTPClient) *DefaultScraperFactory {
	return &DefaultScraperFactory{
		httpClient: httpClient,
	}
}

func (f *DefaultScraperFactory) CreateScraper(scraperType string, state *store.ScraperState) common.Scraper {
	switch scraperType {
	case "Howoge":
		return common.NewBaseScraper(f.httpClient, state, scraperType, howoge.Scrape)
	default:
		return nil
	}
}
