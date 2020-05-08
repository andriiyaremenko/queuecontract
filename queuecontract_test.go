package queuecontract

import (
	"encoding/json"

	"github.com/andriiyaremenko/queuecontract/domain"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-chaincode-go/shimtest"

	//	"github.com/hyperledger/fabric/core/chaincode/shim"
	"testing"
)

type ValueItem struct {
	ItemId string `json:"id"`
	Value  int    `json:"value"`
}

func mapToValueItem(item domain.Item) (ValueItem, bool) {
	v, ok := item.Data["value"]
	if value, convOk := v.(float64); ok && convOk {
		return ValueItem{ItemId: item.Id, Value: int(value)}, true
	}
	return ValueItem{}, false
}

func ActiveFilters(names []string) []domain.Filter {
	var result []domain.Filter
	filters := []domain.Filter{
		domain.Filter{
			Name: "odd",
			F: func(i domain.Item) bool {
				vi, ok := mapToValueItem(i)
				return ok && vi.Value%2 == 0
			},
		},
	}

	for _, f := range filters {
		for _, n := range names {
			if f.Name == n {
				result = append(result, f)
			}
		}
	}
	return result
}

func ActiveSorts(names []string) []domain.Sort {
	var result []domain.Sort
	sorts := []domain.Sort{
		domain.Sort{
			Name: "desc",
			F: func(i1, i2 domain.Item) bool {
				vi1, ok1 := mapToValueItem(i1)
				vi2, ok2 := mapToValueItem(i2)
				return ok1 && ok2 && vi1.Value > vi2.Value
			},
		},
	}

	for _, f := range sorts {
		for _, n := range names {
			if f.Name == n {
				result = append(result, f)
			}
		}
	}
	return result
}

func TestQueueContract(t *testing.T) {
	cc := NewQueueContract(ActiveFilters, ActiveSorts)
	mockStub := shimtest.NewMockStub("Test", cc)
	mockStub.MockInvoke("1", [][]byte{[]byte("Init")})
	r := mockStub.MockInvoke("1", [][]byte{[]byte("Put"), []byte(`{"data":{"value":1}}`)})
	if r.Status != shim.OK {
		t.Errorf("Failed to invoke method: %v", r.GetMessage())
	}
	r = mockStub.MockInvoke("1", [][]byte{[]byte("Peek")})
	if r.Status != shim.OK {
		t.Errorf("Failed to invoke method: %v", r.GetMessage())
	}
	i := new(domain.Item)
	err := json.Unmarshal(r.GetPayload(), i)
	if err != nil {
		t.Errorf("Failed to invoke method: %v", err)
	}
	if v := int(i.Data["value"].(float64)); v != 1 {
		t.Errorf(`Item.Data["value"] = %v; wanted = 1`, v)
	}
	mockStub.MockInvoke("1", [][]byte{[]byte("Put"), []byte(`{"data":{"value":1}}`)})
	mockStub.MockInvoke("1", [][]byte{[]byte("Put"), []byte(`{"data":{"value":2}}`)})
	mockStub.MockInvoke("1", [][]byte{[]byte("Put"), []byte(`{"data":{"value":3}}`)})
	mockStub.MockInvoke("1", [][]byte{[]byte("Put"), []byte(`{"data":{"value":4}}`)})
	mockStub.MockInvoke("1", [][]byte{[]byte("Put"), []byte(`{"data":{"value":5}}`)})
	mockStub.MockInvoke("1", [][]byte{[]byte("Put"), []byte(`{"data":{"value":6}}`)})
	mockStub.MockInvoke("1", [][]byte{[]byte("Put"), []byte(`{"data":{"value":7}}`)})
	mockStub.MockInvoke("1", [][]byte{[]byte("Put"), []byte(`{"data":{"value":8}}`)})
	mockStub.MockInvoke("1", [][]byte{[]byte("Put"), []byte(`{"data":{"value":9}}`)})
	mockStub.MockInvoke("1", [][]byte{[]byte("Put"), []byte(`{"data":{"value":10}}`)})
	mockStub.MockInvoke("1", [][]byte{[]byte("Put"), []byte(`{"data":{"value":11}}`)})
	mockStub.MockInvoke("1", [][]byte{[]byte("Put"), []byte(`{"data":{"value":12}}`)})
	mockStub.MockInvoke("1", [][]byte{[]byte("Put"), []byte(`{"data":{"value":13}}`)})
	mockStub.MockInvoke("1", [][]byte{[]byte("Put"), []byte(`{"data":{"value":14}}`)})
	mockStub.MockInvoke("1", [][]byte{[]byte("Put"), []byte(`{"data":{"value":15}}`)})
	mockStub.MockInvoke("1", [][]byte{[]byte("AddFilter"), []byte(`odd`)})
	r = mockStub.MockInvoke("1", [][]byte{[]byte("Peek")})
	if r.Status != shim.OK {
		t.Errorf("Failed to invoke method: %v", r.GetMessage())
	}
	i = new(domain.Item)
	err = json.Unmarshal(r.GetPayload(), i)
	if err != nil {
		t.Errorf("Failed to invoke method: %v", err)
	}
	if v := int(i.Data["value"].(float64)); v != 2 {
		t.Errorf(`Item.Data["value"] = %v; wanted = 2`, v)
	}
	mockStub.MockInvoke("1", [][]byte{[]byte("RemoveFilter"), []byte(`odd`)})
	r = mockStub.MockInvoke("1", [][]byte{[]byte("Peek")})
	if r.Status != shim.OK {
		t.Errorf("Failed to invoke method: %v", r.GetMessage())
	}
	i = new(domain.Item)
	err = json.Unmarshal(r.GetPayload(), i)
	if err != nil {
		t.Errorf("Failed to invoke method: %v", err)
	}
	if v := int(i.Data["value"].(float64)); v != 1 {
		t.Errorf(`Item.Data["value"] = %v; wanted = 1`, v)
	}
	mockStub.MockInvoke("1", [][]byte{[]byte("AddSort"), []byte(`desc`)})
	mockStub.MockInvoke("1", [][]byte{[]byte("AddFilter"), []byte(`odd`)})
	r = mockStub.MockInvoke("1", [][]byte{[]byte("Peek")})
	if r.Status != shim.OK {
		t.Errorf("Failed to invoke method: %v", r.GetMessage())
	}
	i = new(domain.Item)
	err = json.Unmarshal(r.GetPayload(), i)
	if err != nil {
		t.Errorf("Failed to invoke method: %v", err)
	}
	if v := int(i.Data["value"].(float64)); v != 14 {
		t.Errorf(`Item.Data["value"] = %v; wanted = 14`, v)
	}
	mockStub.MockInvoke("1", [][]byte{[]byte("RemoveFilter"), []byte(`odd`)})
	r = mockStub.MockInvoke("1", [][]byte{[]byte("Peek")})
	if r.Status != shim.OK {
		t.Errorf("Failed to invoke method: %v", r.GetMessage())
	}
	i = new(domain.Item)
	err = json.Unmarshal(r.GetPayload(), i)
	if err != nil {
		t.Errorf("Failed to invoke method: %v", err)
	}
	if v := int(i.Data["value"].(float64)); v != 15 {
		t.Errorf(`Item.Data["value"] = %v; wanted = 15`, v)
	}
	mockStub.MockInvoke("1", [][]byte{[]byte("RemoveSort"), []byte(`desc`)})
	r = mockStub.MockInvoke("1", [][]byte{[]byte("Peek")})
	if r.Status != shim.OK {
		t.Errorf("Failed to invoke method: %v", r.GetMessage())
	}
	i = new(domain.Item)
	err = json.Unmarshal(r.GetPayload(), i)
	if err != nil {
		t.Errorf("Failed to invoke method: %v", err)
	}
	if v := int(i.Data["value"].(float64)); v != 3 {
		t.Errorf(`Item.Data["value"] = %v; wanted = 3`, v)
	}
}
