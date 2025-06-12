package main

import (
	"apartmenthunter/internal/bot"
	"apartmenthunter/internal/config"
	"apartmenthunter/internal/http"
	"apartmenthunter/internal/scraping/common"
	"apartmenthunter/internal/store"
	"apartmenthunter/internal/telegram"
	"context"
	"log"
	"sync"
	"time"
)

var scrapersTypes = []string{
	"Howoge",
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

	httpClient := http.NewClient(5 * time.Second)
	scraperFactory := NewScraperFactory(httpClient)

	startAllScrapers(ctx, scraperFactory)

	select {}
}

func startAllScrapers(ctx context.Context, factory *DefaultScraperFactory) {
	var wg sync.WaitGroup
	for _, scraperType := range scrapersTypes {
		wg.Add(1)
		scraper := factory.CreateScraper(scraperType, store.NewScraperState())
		if scraper == nil {
			log.Printf("unknown scraper type: %s", scraperType)
			wg.Done()
			continue
		}

		// start scraper in its own go routine
		go func(s common.Scraper) {
			defer wg.Done()
			startScraper(ctx, s)
		}(scraper)
	}
}

func startScraper(ctx context.Context, scraper common.Scraper) {
	name := scraper.GetName()
	log.Printf("starting scraper %s", name)

	err := scraper.Scrape(ctx, false)
	if err != nil {
		log.Printf("error during initial scrape for %s: %v", name, err)
	}

	log.Printf("%s scraper initialized", name)

	// start monitoring for new listings
	for {
		select {
		case <-ctx.Done():
			log.Printf("%s scraper stopped", name)
			return
		default:
			time.Sleep(bot.GenerateRandomJitterTime())
			err := scraper.Scrape(ctx, true)
			if err != nil {
				log.Printf("error during initial scrape for %s: %v", name, err)
			}
		}
	}
}
