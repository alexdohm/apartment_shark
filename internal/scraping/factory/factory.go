package factory

import (
	"apartmenthunter/internal/http"
	"apartmenthunter/internal/scraping/common"
	"apartmenthunter/internal/scraping/companies/dewego"
	"apartmenthunter/internal/scraping/companies/gewobag"
	"apartmenthunter/internal/scraping/companies/howoge"
	"apartmenthunter/internal/scraping/companies/stadtundland"
	"apartmenthunter/internal/scraping/companies/wbm"
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
		return common.NewBaseScraper(f.httpClient, state, scraperType, howoge.FetchListings)
	case "Dewego":
		return common.NewBaseScraper(f.httpClient, state, scraperType, dewego.FetchListings)
	case "Gewobag":
		return common.NewBaseScraper(f.httpClient, state, scraperType, gewobag.FetchListings)
	case "StadtUndLand":
		return common.NewBaseScraper(f.httpClient, state, scraperType, stadtundland.FetchListings)
	case "WBM":
		return common.NewBaseScraper(f.httpClient, state, scraperType, wbm.FetchListings)
	default:
		return nil
	}
}
