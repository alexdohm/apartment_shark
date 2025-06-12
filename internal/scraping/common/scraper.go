package common

import (
	"apartmenthunter/internal/bot"
	"apartmenthunter/internal/http"
	"apartmenthunter/internal/store"
	"context"
)

type Scraper interface {
	GetName() string
	Scrape(ctx context.Context, sendTelegram bool) error
}

type ScrapingFunc func(ctx context.Context, scraper *BaseScraper, sendTelegram bool) error

type BaseScraper struct {
	HTTPClient      http.HTTPClient
	HeaderGenerator *bot.HeaderGenerator
	State           *store.ScraperState
	name            string
	scrapingFunc    ScrapingFunc
}

func NewBaseScraper(httpClient http.HTTPClient, state *store.ScraperState, name string, scrapingFunc ScrapingFunc) *BaseScraper {
	return &BaseScraper{
		HTTPClient:      httpClient,
		HeaderGenerator: bot.NewHeaderGenerator(),
		State:           state,
		name:            name,
		scrapingFunc:    scrapingFunc,
	}
}

func (b *BaseScraper) GetName() string {
	return b.name
}

func (b *BaseScraper) Scrape(ctx context.Context, sendTelegram bool) error {
	return b.scrapingFunc(ctx, b, sendTelegram)
}
