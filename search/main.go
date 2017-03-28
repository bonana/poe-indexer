// File to search the PoE stash river for items

package main

import (
	"bufio"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/bonana/poe-indexer/api"
)

// Function to read item filters
func readFilter(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// Function to match type to line in our filter
func matchFilter(s string, list []string) bool {
	for _, b := range list {
		if s == b {
			return true
		}
	}
	return false
}

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

func processStash(stash *api.Stash, filters []string) {
	for _, item := range stash.Items {
		if matchFilter(item.Type, filters) {
			// Some items miss names, use type instead.
			name := item.Type

			// If item has specific name use that instead.
			if item.Name != "" {
				name = item.Name
			}

			log.Printf("Found matching item: name = %v, account = %v, league = %v, note = %v, tab = %v, character = %v", name, stash.AccountName, item.League, item.Note, stash.Label, stash.LastCharacterName)
		}
	}
}

func main() {
	log.Printf("loading item filters")
	filters, err := readFilter("poe.filter")
	if err != nil {
		log.Fatal(err)
	}

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
			processStash(&stash, filters)
		}
	}
}
