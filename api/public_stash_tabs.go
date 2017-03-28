// Structure for stash tab API

package api

type PublicStashTabs struct {
	NextChangeId string  `json:"next_change_id"`
	Stashes      []Stash `json:"stashes"`
}
