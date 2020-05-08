package domain

import (
	_ "encoding/json"
	"github.com/hyperledger/fabric-chaincode-go/shim"
)

type Item struct {
	Id   string                 `json:"id"`
	Data map[string]interface{} `json:"data"`
}

type QueueContext struct {
	Items   []Item
	Filters []string
	Sorts   []string
}

type Sort struct {
	F    func(Item, Item) bool
	Name string
}

type ActiveSorts func(names []string) []Sort

type Filter struct {
	F    func(Item) bool
	Name string
}

type ActiveFilters func(names []string) []Filter

type QueueContract interface {
	shim.Chaincode
	Put(shim.ChaincodeStubInterface, Item) error
	Peek(shim.ChaincodeStubInterface) (*Item, error)
	AddSort(shim.ChaincodeStubInterface, string) error
	RemoveSort(shim.ChaincodeStubInterface, string) error
	AddFilter(shim.ChaincodeStubInterface, string) error
	RemoveFilter(shim.ChaincodeStubInterface, string) error
}
