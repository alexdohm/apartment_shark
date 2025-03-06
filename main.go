package main

import (
	"apartmenthunter/config"
	"apartmenthunter/listings"
	"apartmenthunter/scraping"
	"log"
	"time"
)

func main() {
	log.Println("starting project")
	startGewobag()
	startWbm()
	startHowoge()
	startDewego()

	select {}
}

func startDewego() {
	//not getting very many listings
	dewegoListings, err := listings.LoadListings(config.DewegoFile)
	if err != nil {
		log.Printf("Could not load gewobag listings: %v", err)
	}

	go func() {
		for {
			scraping.CheckDewego(dewegoListings)
			time.Sleep(config.TimeBetweenCalls * time.Second)
		}
	}()
}

func startHowoge() {
	howogeListings, err := listings.LoadListings(config.HowogeFile)
	if err != nil {
		log.Printf("Could not load gewobag listings: %v", err)
	}

	go func() {
		for {
			scraping.CheckHowoge(howogeListings)
			time.Sleep(config.TimeBetweenCalls * time.Second)
		}
	}()
}

func startGewobag() {
	gewobagListings, err := listings.LoadListings(config.GewobagFile)
	if err != nil {
		log.Printf("Could not load gewobag listings: %v", err)
	}

	go func() {
		for {
			scraping.CheckGewobag(gewobagListings)
			time.Sleep(config.TimeBetweenCalls * time.Second)
		}
	}()
}

func startWbm() {
	wbmListings, err := listings.LoadListings(config.WbmFile)
	if err != nil {
		log.Printf("Could not load wbm listings: %v", err)
	}

	go func() {
		for {
			scraping.CheckWbm(wbmListings)
			time.Sleep(config.TimeBetweenCalls * time.Second)
		}
	}()
}
