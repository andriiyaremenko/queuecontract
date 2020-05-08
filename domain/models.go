package domain

import (
	_ "encoding/json"
)

type Item struct {
	Id   string                 `json:"id"`
	Data map[string]interface{} `json:"data"`
}

type QueueContext struct {
	Items   []Item   `json:"items"`
	Filters []string `json:"filters"`
	Sorts   []string `json:"sorts"`
}

type Sort struct {
	F    func(Item, Item) bool
	Name string
}

type Filter struct {
	F    func(Item) bool
	Name string
}
