// Structures to subscribe to the public stash tab river

package api

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"
)

// Struct to handle results from the PoE API
type PublicStashTabSubscriptionResult struct {
	PublicStashTabs *PublicStashTabs
	Error           error
}

// Struct to handle subcriptions to the PoE API
type PublicStashTabSubscription struct {
	Channel      chan PublicStashTabSubscriptionResult
	closeChannel chan bool
	host         string
}

// Function to open a subscription to the official PoE API with the given change ID.
// To subscribe from the beginning of river just pass an empty string.
func OpenPublicStashTabSubscription(firstChangeId string) *PublicStashTabSubscription {
	ret := &PublicStashTabSubscription{
		Channel:      make(chan PublicStashTabSubscriptionResult),
		closeChannel: make(chan bool),
		host:         "www.pathofexile.com",
	}
	go ret.run(firstChangeId)
	return ret
}

func (s *PublicStashTabSubscription) Close() {
	s.closeChannel <- true
}

// Function to start subscribing to the PoE API and listen to updates from the river
func (s *PublicStashTabSubscription) run(firstChangeId string) {
	defer close(s.Channel)

	nextChangeId := firstChangeId

	const requestInterval = time.Second
	var lastRequestTime time.Time

	// Start listening for updates
	for {
		waitTime := requestInterval - time.Now().Sub(lastRequestTime)
		if waitTime > 0 {
			time.Sleep(waitTime)
		}

		select {
		case <-s.closeChannel:
			return
		default:
			lastRequestTime = time.Now()
			response, err := http.Get("https://" + s.host + "/api/public-stash-tabs?id=" + url.QueryEscape(nextChangeId))
			if err != nil {
				s.Channel <- PublicStashTabSubscriptionResult{
					Error: err,
				}
				continue
			}

			tabs := new(PublicStashTabs)
			decoder := json.NewDecoder(response.Body)
			err = decoder.Decode(tabs)
			if err != nil {
				s.Channel <- PublicStashTabSubscriptionResult{
					Error: err,
				}
				continue
			}

			nextChangeId = tabs.NextChangeId

			if len(tabs.Stashes) > 0 {
				s.Channel <- PublicStashTabSubscriptionResult{
					PublicStashTabs: tabs,
				}
			}
		}
	}
}
