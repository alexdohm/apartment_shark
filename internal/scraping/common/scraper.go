package common

import (
	"apartmenthunter/internal/bot"
	"apartmenthunter/internal/http"
	"apartmenthunter/internal/store"
	"context"
	"fmt"
)

type Scraper interface {
	GetName() string
	Scrape(ctx context.Context) ([]Listing, error)
	GetState() *store.ScraperState
}

type ScrapingFunc func(ctx context.Context, scraper *BaseScraper) ([]Listing, error)

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

func (b *BaseScraper) Scrape(ctx context.Context) ([]Listing, error) {
	listings, err := b.scrapingFunc(ctx, b)
	if err != nil {
		return nil, fmt.Errorf("fetching %s listings: %w", b.name, err)
	}
	return listings, nil
}

func (b *BaseScraper) GetState() *store.ScraperState {
	return b.State
}
