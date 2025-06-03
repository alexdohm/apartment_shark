package main

import (
	"apartmenthunter/internal/bot"
	"apartmenthunter/internal/config"
	"apartmenthunter/internal/scraping"
	"apartmenthunter/internal/store"
	"apartmenthunter/internal/telegram"
	"context"
	"log"
	"sync"
	"time"
)

type ScraperConfig struct {
	Name      string
	State     *store.ScraperState
	CheckFunc func(context.Context, *store.ScraperState, bool)
}

var scrapers = []ScraperConfig{
	{"StadtUndLand", store.NewScraperState(), scraping.CheckStadtUndLand},
	{"Dewego", store.NewScraperState(), scraping.CheckDewego},
	{"Howoge", store.NewScraperState(), scraping.CheckHowoge},
	{"Gewobag", store.NewScraperState(), scraping.CheckGewobag},
	{"WBM", store.NewScraperState(), scraping.CheckWbm},
}

func main() {
	log.Println("starting apartment project")

	if err := telegram.Init(config.BaseURL, config.BotToken, config.ChatID); err != nil {
		log.Fatalf("error initializing telegram: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	htmlMsg := "<b>Apartment Hunter</b> is <i>running...</i>"
	if err := telegram.SendStartup(ctx, htmlMsg); err != nil {
		log.Fatalf("error sending startup message: %v", err)
	}

	startAllScrapers(ctx)

	select {}
}

func startAllScrapers(ctx context.Context) {
	var wg sync.WaitGroup
	for _, scraper := range scrapers {
		wg.Add(1)
		go func(s ScraperConfig) {
			defer wg.Done()
			startScraper(ctx, s)
		}(scraper)
	}
}

func startScraper(ctx context.Context, conf ScraperConfig) {
	log.Printf("starting scraper %s", conf.Name)

	initCtx, initCancel := context.WithTimeout(ctx, 5*time.Second)
	conf.CheckFunc(initCtx, conf.State, false)
	initCancel()

	log.Printf("%s scraper initialized", conf.Name)

	// start monitoring for new listings
	for {
		select {
		case <-ctx.Done():
			log.Printf("%s scraper stopped", conf.Name)
			return
		default:
			time.Sleep(bot.GenerateRandomJitterTime())
			opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			conf.CheckFunc(opCtx, conf.State, true)
			cancel()
		}
	}
}
