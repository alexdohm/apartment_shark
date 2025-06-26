package factory

import (
	"apartmenthunter/internal/http"
	"apartmenthunter/internal/scraping/common"
	"apartmenthunter/internal/scraping/dewego"
	"apartmenthunter/internal/scraping/gewobag"
	"apartmenthunter/internal/scraping/howoge"
	"apartmenthunter/internal/scraping/stadtundland"
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
	case "Dewego":
		return common.NewBaseScraper(f.httpClient, state, scraperType, dewego.Scrape)
	case "Gewobag":
		return common.NewBaseScraper(f.httpClient, state, scraperType, gewobag.Scrape)
	case "StadtUndLand":
		return common.NewBaseScraper(f.httpClient, state, scraperType, stadtundland.Scrape)
	default:
		return nil
	}
}
