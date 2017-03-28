// File to search the PoE stash river for items

package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/bonana/poe-indexer/api"
)

// To make sure we don't go through several terabytes of data we should grab a recent ID. This is something we can get from the helpful poe.ninja.
func getRecentChangeId() (string, error) {
	resp, err := http.Get("http://api.poe.ninja/api/Data/GetStats")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var stats struct {
		// poe.ninja returns a bit more data than this but that's irrelevant for us
		NextChangeId string `json:"nextChangeId"`
	}
	if err := json.Unmarshal(body, &stats); err != nil {
		return "", err
	}

	return stats.NextChangeId, nil
}

func processStash(stash *api.Stash) {
	for _, item := range stash.Items {
		if item.Type == "Ancient Reliquary Key" {
			log.Printf("Ancient Reliquary Key: account = %v, league = %v, note = %v, tab = %v", stash.accountName, item.League, item.Note, stash.Label)
		}
	}
}

func main() {
	log.Printf("requesting a recent change id from poe.ninja...")
	recentChangeId, err := getRecentChangeId()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("starting with change id %v", recentChangeId)

	subscription := api.OpenPublicStashTabSubscription(recentChangeId)

	// Handle exits gracefully
	go func() {
		ch := make(chan os.Signal)
		signal.Notify(ch, os.Interrupt)
		<-ch
		log.Printf("shutting down")
		subscription.Close()
	}()

	// Loop
	for result := range subscription.Channel {
		if result.Error != nil {
			log.Printf("error: %v", result.Error.Error())
			continue
		}
		for _, stash := range result.PublicStashTabs.Stashes {
			processStash(&stash)
		}
	}
}
