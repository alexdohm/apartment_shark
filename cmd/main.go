package main

import (
	"apartmenthunter/internal/bot"
	"apartmenthunter/internal/config"
	"apartmenthunter/internal/http"
	"apartmenthunter/internal/scraping/common"
	"apartmenthunter/internal/scraping/factory"
	"apartmenthunter/internal/store"
	"apartmenthunter/internal/telegram"
	"context"
	"log"
	"sync"
	"time"
)

var scrapersTypes = []string{
	"Howoge",
	//"Dewego",
	//"Gewobag",
	//"StadtUndLand",
	//"WBM",
}

func main() {
	log.Println("starting apartment project")

	telegramClient, err := telegram.NewClient(config.BaseURL, config.BotToken, config.ChatID)
	if err != nil {
		log.Fatalf("error initializing telegram client: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	htmlMsg := "<b>Apartment Hunter</b> is <i>running...</i>"
	if err := telegramClient.SendStartup(ctx, htmlMsg); err != nil {
		log.Fatalf("error sending startup message: %v", err)
	}

	httpClient := http.NewClient(5 * time.Second)
	scraperFactory := factory.NewScraperFactory(httpClient)

	startAllScrapers(ctx, scraperFactory, telegramClient)

	select {}
}

func startAllScrapers(ctx context.Context, factory *factory.DefaultScraperFactory, client *telegram.Client) {
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
			startScraper(ctx, s, client)
		}(scraper)
	}
}

func startScraper(ctx context.Context, scraper common.Scraper, client *telegram.Client) {
	name := scraper.GetName()
	state := scraper.GetState()
	log.Printf("[%s] starting scraper", name)

	// Initial scrape without notifications - mark existing listings as seen
	initialListings, err := scraper.Scrape(ctx)
	if err != nil {
		log.Printf("[%s] error during initial scrape: %v", name, err)
	} else {
		for _, listing := range initialListings {
			log.Printf("[%s] Storing initial listing: %s", name, listing.ID)
			state.MarkAsSeen(listing.ID)
		}
	}

	log.Printf("[%s] scraper store initialized", name)

	// start monitoring for new listings
	for {
		select {
		case <-ctx.Done():
			log.Printf("[%s] scraper stopped", name)
			return
		default:
			time.Sleep(bot.GenerateRandomJitterTime())

			listings, err := scraper.Scrape(ctx)
			if err != nil {
				log.Printf("[%s] Error during scrape: %v", name, err)
				continue
			}

			// Check for new listings and send notifications
			for _, listing := range listings {
				if !state.Exists(listing.ID) {
					log.Printf("[%s] New listing: %s", name, listing.ID)
					state.MarkAsSeen(listing.ID)

					// Convert to telegram format and send
					telegramInfo := listing.ToTelegramInfo()
					if err := client.SendListing(ctx, telegramInfo); err != nil {
						log.Printf("[%s] Failed to send notification: %v", listing.ID, err)
					}
				}
			}
		}
	}
}
