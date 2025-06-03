package main

import (
	"apartmenthunter/internal/bot"
	"apartmenthunter/internal/config"
	"apartmenthunter/internal/scraping"
	"apartmenthunter/internal/store"
	"apartmenthunter/internal/telegram"
	"context"
	"log"
	"time"
)

var (
	dewegoState       = store.NewScraperState()
	howogeState       = store.NewScraperState()
	gewobagState      = store.NewScraperState()
	wbmState          = store.NewScraperState()
	stadtUndLandState = store.NewScraperState()
)

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

	// load json config
	//startGewobag(ctx)
	//startDewego(ctx)
	//startHowoge(ctx)
	//startWbm(ctx)
	startStadtUndLand(ctx)

	select {}
}

func startStadtUndLand(ctx context.Context) {
	initCtx, initCancel := context.WithTimeout(ctx, 5*time.Second)
	scraping.CheckStadtUndLand(initCtx, stadtUndLandState, false)
	initCancel()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
				scraping.CheckStadtUndLand(opCtx, stadtUndLandState, true)
				cancel()
				time.Sleep(bot.GenerateRandomJitterTime())
			}
		}
	}()
}

//func startDewego() {
//	scraping.CheckDewego(dewegoState, false)
//	go func() {
//		for {
//			time.Sleep(bot.GenerateRandomJitterTime())
//			scraping.CheckDewego(dewegoState, true)
//		}
//	}()
//}
//
//func startHowoge() {
//	scraping.CheckHowoge(howogeState, false)
//	go func() {
//		for {
//			time.Sleep(bot.GenerateRandomJitterTime())
//			scraping.CheckHowoge(howogeState, true)
//		}
//	}()
//}
//
//func startGewobag() {
//	scraping.CheckGewobag(gewobagState, false)
//	go func() {
//		for {
//			time.Sleep(bot.GenerateRandomJitterTime())
//			scraping.CheckGewobag(gewobagState, true)
//		}
//	}()
//}
//
//func startWbm() {
//	go scraping.CheckWbm(wbmState, false)
//	go func() {
//		for {
//			time.Sleep(bot.GenerateRandomJitterTime())
//			scraping.CheckWbm(wbmState, true)
//		}
//	}()
//}
