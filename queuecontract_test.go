package queuecontract

import (
	"encoding/json"

	"github.com/andriiyaremenko/queuecontract/domain"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-chaincode-go/shimtest"

	pb "github.com/hyperledger/fabric-protos-go/peer"
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

func initQC() (*shimtest.MockStub, pb.Response) {
	cc := NewQueueChaincode(ActiveFilters, ActiveSorts)
	mockStub := shimtest.NewMockStub("Test", cc)
	return mockStub, mockStub.MockInvoke("1", [][]byte{[]byte("Init")})
}

func TestQueueContractInit(t *testing.T) {
	_, r := initQC()
	if r.Status != shim.OK {
		t.Errorf("Failed to invoke method: %v", r.GetMessage())
	}
}

func TestQueueContractPut(t *testing.T) {
	mockStub, _ := initQC()
	r := mockStub.MockInvoke("1", [][]byte{[]byte("Put"), []byte(`{"data":{"value":1}}`)})
	if r.Status != shim.OK {
		t.Errorf("Failed to invoke method: %v", r.GetMessage())
	}
}

func checkR(t *testing.T, r pb.Response, want int) {
	if r.Status != shim.OK {
		t.Errorf("Failed to invoke method: %v", r.GetMessage())
	}
	i := new(domain.Item)
	err := json.Unmarshal(r.GetPayload(), i)
	if err != nil {
		t.Errorf("Failed to invoke method: %v", err)
	}
	if v := int(i.Data["value"].(float64)); v != want {
		t.Errorf(`Item.Data["value"] = %v; wanted = %v`, v, want)
	}
}

func TestQueueContractPeek(t *testing.T) {
	mockStub, _ := initQC()
	mockStub.MockInvoke("1", [][]byte{[]byte("Put"), []byte(`{"data":{"value":1}}`)})
	r := mockStub.MockInvoke("1", [][]byte{[]byte("Peek")})
	checkR(t, r, 1)
}

func TestQueueContractSort(t *testing.T) {
	mockStub, _ := initQC()
	r := mockStub.MockInvoke(
		"1",
		[][]byte{
			[]byte("Put"),
			[]byte(`{"data":{"value":1}}`),
			[]byte(`{"data":{"value":2}}`),
			[]byte(`{"data":{"value":3}}`),
		},
	)
	if r.Status != shim.OK {
		t.Errorf("Failed to invoke method: %v", r.GetMessage())
	}
	mockStub.MockInvoke("1", [][]byte{[]byte("AddSort"), []byte(`desc`)})
	r = mockStub.MockInvoke("1", [][]byte{[]byte("Peek")})
	checkR(t, r, 3)
	mockStub.MockInvoke("1", [][]byte{[]byte("RemoveSort"), []byte(`desc`)})
	r = mockStub.MockInvoke("1", [][]byte{[]byte("Peek")})
	checkR(t, r, 1)
}

func TestQueueContractFilter(t *testing.T) {
	mockStub, _ := initQC()
	r := mockStub.MockInvoke(
		"1",
		[][]byte{
			[]byte("Put"),
			[]byte(`{"data":{"value":1}}`),
			[]byte(`{"data":{"value":2}}`),
			[]byte(`{"data":{"value":3}}`),
		},
	)
	if r.Status != shim.OK {
		t.Errorf("Failed to invoke method: %v", r.GetMessage())
	}
	mockStub.MockInvoke("1", [][]byte{[]byte("AddFilter"), []byte(`odd`)})
	r = mockStub.MockInvoke("1", [][]byte{[]byte("Peek")})
	checkR(t, r, 2)
	mockStub.MockInvoke("1", [][]byte{[]byte("RemoveFilter"), []byte(`odd`)})
	r = mockStub.MockInvoke("1", [][]byte{[]byte("Peek")})
	checkR(t, r, 1)
}

func TestQueueContractFilterSort(t *testing.T) {
	mockStub, _ := initQC()
	r := mockStub.MockInvoke(
		"1",
		[][]byte{
			[]byte("Put"),
			[]byte(`{"data":{"value":1}}`),
			[]byte(`{"data":{"value":2}}`),
			[]byte(`{"data":{"value":3}}`),
			[]byte(`{"data":{"value":4}}`),
			[]byte(`{"data":{"value":5}}`),
			[]byte(`{"data":{"value":6}}`),
			[]byte(`{"data":{"value":7}}`),
		},
	)
	if r.Status != shim.OK {
		t.Errorf("Failed to invoke method: %v", r.GetMessage())
	}
	mockStub.MockInvoke("1", [][]byte{[]byte("AddFilter"), []byte(`odd`)})
	mockStub.MockInvoke("1", [][]byte{[]byte("AddSort"), []byte(`desc`)})
	r = mockStub.MockInvoke("1", [][]byte{[]byte("Peek")})
	checkR(t, r, 6)
	mockStub.MockInvoke("1", [][]byte{[]byte("RemoveFilter"), []byte(`odd`)})
	r = mockStub.MockInvoke("1", [][]byte{[]byte("Peek")})
	checkR(t, r, 7)
	mockStub.MockInvoke("1", [][]byte{[]byte("RemoveSort"), []byte(`desc`)})
	r = mockStub.MockInvoke("1", [][]byte{[]byte("Peek")})
	checkR(t, r, 1)
}
