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
	//startDewego()

	select {}
}

// func startDewego() {} {
//
// }
func startGewobag() {
	gewobagListings, err := listings.LoadListings(config.GewobagFile)
	if err != nil {
		log.Fatalf("Could not load gewobag listings: %v", err)
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
		log.Fatalf("Could not load wbm listings: %v", err)
	}

	go func() {
		for {
			scraping.CheckWbm(wbmListings)
			time.Sleep(config.TimeBetweenCalls * time.Second)
		}
	}()
}
