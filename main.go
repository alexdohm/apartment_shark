package main

import (
	"apartmenthunter/bot"
	"apartmenthunter/scraping"
	"apartmenthunter/store"
	"apartmenthunter/telegram"
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
	log.Println("starting project")
	htmlMsg := "<b>Apartment Hunter</b> is <i>starting...</i>"
	telegram.SendTelegramMessage(htmlMsg)

	//startGewobag()
	//startDewego()
	startHowoge()
	//startWbm()
	//startStadtUndLand()

	select {}
}

func startStadtUndLand() {
	scraping.CheckStadtUndLand(stadtUndLandState, false)
	go func() {
		for {
			time.Sleep(bot.GenerateRandomJitterTime())
			scraping.CheckStadtUndLand(stadtUndLandState, true)
		}
	}()
}

func startDewego() {
	scraping.CheckDewego(dewegoState, false)
	go func() {
		for {
			time.Sleep(bot.GenerateRandomJitterTime())
			scraping.CheckDewego(dewegoState, true)
		}
	}()
}

func startHowoge() {
	scraping.CheckHowoge(howogeState, false)
	go func() {
		for {
			time.Sleep(bot.GenerateRandomJitterTime())
			scraping.CheckHowoge(howogeState, true)
		}
	}()
}

func startGewobag() {
	scraping.CheckGewobag(gewobagState, false)
	go func() {
		for {
			time.Sleep(bot.GenerateRandomJitterTime())
			scraping.CheckGewobag(gewobagState, true)
		}
	}()
}

func startWbm() {
	go scraping.CheckWbm(wbmState, false)
	go func() {
		for {
			time.Sleep(bot.GenerateRandomJitterTime())
			scraping.CheckWbm(wbmState, true)
		}
	}()
}
