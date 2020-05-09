package domain

import (
	_ "encoding/json"
)

type Payload map[string]interface{}

type Item struct {
	Id   string  `json:"id"`
	Data Payload `json:"data"`
}

type QueueContext struct {
	Items []Item `json:"items"`
}

type FilterRequest map[string][]interface{}

type SortRequest map[string][]interface{}

type Sort struct {
	F    func(Item, Item, ...interface{}) bool
	Name string
}

type Filter struct {
	F    func(Item, ...interface{}) bool
	Name string
}

type Validator struct {
	F    func(Payload) error
	Name string
}
