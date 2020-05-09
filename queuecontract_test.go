package queuecontract

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/andriiyaremenko/queuecontract/domain"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-chaincode-go/shimtest"

	"testing"

	pb "github.com/hyperledger/fabric-protos-go/peer"
)

type ValueItem struct {
	ItemId string `json:"id"`
	Value  int    `json:"value"`
}

func validators() ([]domain.Validator, error) {
	return []domain.Validator{
		domain.Validator{
			Name: "ValueItem",
			F: func(p domain.Payload) error {
				if len(p) == 0 {
					return nil
				}
				v, ok := p["value"]
				if _, convOk := v.(float64); ok && convOk {
					return nil
				}
				return errors.New(`"value" field is missing`)
			},
		},
	}, nil
}

func mapToValueItem(item domain.Item) (ValueItem, bool) {
	v, ok := item.Data["value"]
	if !ok {
		return ValueItem{ItemId: item.Id}, true
	}
	if value, convOk := v.(float64); convOk {
		return ValueItem{ItemId: item.Id, Value: int(value)}, true
	}
	return ValueItem{}, false
}

func activeFilters(filterRequest domain.FilterRequest) ([]domain.Filter, error) {
	var result []domain.Filter
	filters := []domain.Filter{
		domain.Filter{
			Name: "odd",
			F: func(i domain.Item, args ...interface{}) bool {
				vi, ok := mapToValueItem(i)
				return ok && vi.Value%2 == 0
			},
		},
		domain.Filter{
			Name: "byIds",
			F: func(i domain.Item, args ...interface{}) bool {
				vi, ok := mapToValueItem(i)
				if !ok {
					return false
				}
				for _, arg := range args {
					if vi.ItemId == arg.(string) {
						return true
					}
				}
				return false
			},
		},
	}

	for _, f := range filters {
		for n, _ := range filterRequest {
			if f.Name == n {
				result = append(result, f)
			}
		}
	}
	return result, nil
}

func activeSorts(sortRequest domain.SortRequest) ([]domain.Sort, error) {
	var result []domain.Sort
	sorts := []domain.Sort{
		domain.Sort{
			Name: "desc",
			F: func(i1, i2 domain.Item, args ...interface{}) bool {
				vi1, ok1 := mapToValueItem(i1)
				vi2, ok2 := mapToValueItem(i2)
				return ok1 && ok2 && vi1.Value > vi2.Value
			},
		},
	}

	for _, f := range sorts {
		for n, _ := range sortRequest {
			if f.Name == n {
				result = append(result, f)
			}
		}
	}
	return result, nil
}

func initQC() (*shimtest.MockStub, pb.Response) {
	cc := NewQueueChaincode(activeFilters, activeSorts, validators)
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
		t.Errorf("Bad response: %v", err)
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
	r = mockStub.MockInvoke("1", [][]byte{[]byte("Peek"), []byte(`{"desc": []}`)})
	checkR(t, r, 3)
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
	r = mockStub.MockInvoke("1", [][]byte{[]byte("Peek"), []byte(nil), []byte(`{"odd": []}`)})
	checkR(t, r, 2)
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
	r = mockStub.MockInvoke("1", [][]byte{[]byte("Peek"), []byte(`{"desc": []}`), []byte(`{"odd": []}`)})
	checkR(t, r, 6)
	r = mockStub.MockInvoke("1", [][]byte{[]byte("Peek")})
	checkR(t, r, 1)
}

func TestQueueContractShouldValidate(t *testing.T) {
	mockStub, _ := initQC()
	r := mockStub.MockInvoke(
		"1",
		[][]byte{
			[]byte("Put"),
			[]byte(`{"data":{"value":1}}`),
			[]byte(`{"data":{"badData":2}}`),
			[]byte(`{"data":{"badData":3}}`),
			[]byte(`{"data":{"badData":4}}`),
			[]byte(`{"data":{"badData":5}}`),
			[]byte(`{"data":{"badData":6}}`),
			[]byte(`{"data":{"badData":7}}`),
		},
	)
	if r.Status != shim.ERROR {
		t.Errorf("Validation works incorrectly: expected Response.Status to be %d, got %d", shim.ERROR, r.Status)
	}
	t.Logf("Validation works correctly: got validation errors: %v", r.GetMessage())
}

func TestQueueContractUpdate(t *testing.T) {
	mockStub, _ := initQC()
	r := mockStub.MockInvoke(
		"1",
		[][]byte{
			[]byte("Put"),
			[]byte(`{"data":{}}`),
			[]byte(`{"data":{}}`),
			[]byte(`{"data":{}}`),
			[]byte(`{"data":{}}`),
		},
	)
	if r.Status != shim.OK {
		t.Errorf("Failed to invoke method: %v", r.GetMessage())
	}
	ids := new([]string)
	raw := r.GetPayload()
	err := json.Unmarshal(raw, ids)
	if err != nil {
		t.Errorf("Bad response: %v", err)
	}
	filterArg1 := (*ids)[:2]
	rawFilterArg1, err := json.Marshal(filterArg1)
	r = mockStub.MockInvoke(
		"1",
		[][]byte{
			[]byte("Update"),
			[]byte(`{"value": 25}`),
			[]byte(fmt.Sprintf(`{"byIds": %s}`, rawFilterArg1)),
		},
	)
	if r.Status != shim.OK {
		t.Errorf("Failed to invoke method: %v", r.GetMessage())
	}
	filterArg2 := (*ids)[2:]
	rawFilterArg2, err := json.Marshal(filterArg2)
	r = mockStub.MockInvoke(
		"1",
		[][]byte{
			[]byte("Update"),
			[]byte(`{"value": 3}`),
			[]byte(fmt.Sprintf(`{"byIds": %s}`, rawFilterArg2)),
		},
	)
	if r.Status != shim.OK {
		t.Errorf("Failed to invoke method: %v", r.GetMessage())
	}
	r = mockStub.MockInvoke(
		"1",
		[][]byte{
			[]byte("Peek"),
			[]byte(nil),
			[]byte(fmt.Sprintf(`{"byIds": %s}`, rawFilterArg1)),
		},
	)
	if r.Status != shim.OK {
		t.Errorf("Failed to invoke method: %v", r.GetMessage())
	}
	checkR(t, r, 25)
	r = mockStub.MockInvoke(
		"1",
		[][]byte{
			[]byte("Peek"),
			[]byte(nil),
			[]byte(fmt.Sprintf(`{"byIds": %s}`, rawFilterArg2)),
		},
	)
	if r.Status != shim.OK {
		t.Errorf("Failed to invoke method: %v", r.GetMessage())
	}
	checkR(t, r, 3)
}
