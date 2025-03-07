package main

import (
	"apartmenthunter/config"
	"apartmenthunter/listings"
	"apartmenthunter/scraping"
	"apartmenthunter/telegram"
	"log"
	"time"
)

var (
	dewegoState       = listings.NewScraperState()
	howogeState       = listings.NewScraperState()
	gewobagState      = listings.NewScraperState()
	wbmState          = listings.NewScraperState()
	stadtUndLandState = listings.NewScraperState()
)

func main() {
	log.Println("starting project")
	htmlMsg := "<b>Apartment Hunter</b> is <i>starting...</i>"
	telegram.SendTelegramMessage(htmlMsg)

	startDewego()
	startHowoge()
	startGewobag()
	startWbm()
	startStadtUndLand()

	select {}
}

func startStadtUndLand() {
	scraping.CheckStadtUndLand(stadtUndLandState, false)
	go func() {
		for {
			time.Sleep(config.TimeBetweenCalls * time.Second)
			scraping.CheckStadtUndLand(stadtUndLandState, true)
		}
	}()
}

func startDewego() {
	scraping.CheckDewego(dewegoState, false)
	go func() {
		for {
			time.Sleep(config.TimeBetweenCalls * time.Second)
			scraping.CheckDewego(dewegoState, true)
		}
	}()
}

func startHowoge() {
	scraping.CheckHowoge(howogeState, false)
	go func() {
		for {
			time.Sleep(config.TimeBetweenCalls * time.Second)
			scraping.CheckHowoge(howogeState, true)
		}
	}()
}

func startGewobag() {
	scraping.CheckGewobag(gewobagState, false)
	go func() {
		for {
			time.Sleep(config.TimeBetweenCalls * time.Second)
			scraping.CheckGewobag(gewobagState, true)
		}
	}()
}

func startWbm() {
	go scraping.CheckWbm(wbmState, false)
	go func() {
		for {
			time.Sleep(config.TimeBetweenCalls * time.Second)
			scraping.CheckWbm(wbmState, true)
		}
	}()
}
