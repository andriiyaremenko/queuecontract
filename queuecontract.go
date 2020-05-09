package queuecontract

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/andriiyaremenko/queuecontract/domain"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

type queueChaincode struct {
	c domain.QueueContract
}

func NewQueueChaincode(
	activeFilters domain.ActiveFilters, activeSorts domain.ActiveSorts, validators domain.Validators,
) shim.Chaincode {
	return &queueChaincode{c: domain.NewContract(activeFilters, activeSorts, validators)}
}

func (q *queueChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	q.c.Init(stub, new(domain.QueueContext))
	return shim.Success(nil)
}

func (q *queueChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	switch function {
	case "Init":
		return q.Init(stub)
	case "Put":
		return q.put(stub, args)
	case "Peek":
		return q.peek(stub, args)
	case "Update":
		return q.update(stub, args)
	default:
		return shim.Error("NewQueueChaincode(...).Invoke: invalid invoke function name.")
	}
}

func (q *queueChaincode) put(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) == 0 {
		return shim.Error("NewQueueChaincode(...).Invoke: put: missing arguments")
	}
	var sb strings.Builder
	sb.WriteString("NewQueueChaincode(...).Invoke: [")
	var items []domain.Item
	hasError := false
	for _, raw := range args {
		item := new(domain.Item)
		err := json.Unmarshal([]byte(raw), item)
		if err == nil {
			items = append(items, *item)
			continue
		}
		if hasError {
			sb.WriteString(", ")
		}
		hasError = true
		sb.WriteString(err.Error())
	}
	if hasError {
		sb.WriteString("]")
		return shim.Error(sb.String())
	}
	ids, err := q.c.Put(stub, items...)
	if err != nil {
		return shim.Error(fmt.Sprintf("NewQueueChaincode(...).Invoke: %v", err))
	}
	resp, err := json.Marshal(ids)
	if err != nil {
		return shim.Error(fmt.Sprintf("NewQueueChaincode(...).Invoke: %v", err))
	}
	return shim.Success(resp)
}

func (q *queueChaincode) peek(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) > 2 {
		return shim.Error("NewQueueChaincode(...).Invoke: peek: to much arguments")
	}
	sort, err := getSort(args)
	if err != nil {
		return shim.Error(fmt.Sprintf("NewQueueChaincode(...).Invoke: peek: %v", err))
	}
	filter, err := getFilter(args)
	if err != nil {
		return shim.Error(fmt.Sprintf("NewQueueChaincode(...).Invoke: peek: %v", err))
	}
	item, err := q.c.Peek(stub, sort, filter)
	if err != nil {
		return shim.Error(fmt.Sprintf("NewQueueChaincode(...).Invoke: %v", err))
	}
	raw, err := json.Marshal(item)
	if err != nil {
		return shim.Error(fmt.Sprintf("NewQueueChaincode(...).Invoke: %v", err))
	}
	return shim.Success(raw)
}

func (q *queueChaincode) update(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 2 {
		return shim.Error("NewQueueChaincode(...).Invoke: update: missing arguments")
	}
	payload := new(domain.Payload)
	err := json.Unmarshal([]byte(args[0]), payload)
	if err != nil {
		return shim.Error(fmt.Sprintf("NewQueueChaincode(...).Invoke: update: %v", err))
	}
	filter, err := getFilter(args)
	if err != nil {
		return shim.Error(fmt.Sprintf("NewQueueChaincode(...).Invoke: update: %v", err))
	}
	err = q.c.Update(stub, *payload, filter)
	if err != nil {
		return shim.Error(fmt.Sprintf("NewQueueChaincode(...).Invoke: update: %v", err))
	}
	return shim.Success(nil)
}

func getSort(args []string) (domain.SortRequest, error) {
	sort := new(domain.SortRequest)
	if len(args) < 1 || len(args[0]) == 0 {
		return *sort, nil
	}
	err := json.Unmarshal([]byte(args[0]), sort)
	if err != nil {
		return nil, err
	}
	return *sort, nil
}

func getFilter(args []string) (domain.FilterRequest, error) {
	filter := new(domain.FilterRequest)
	if len(args) < 2 || len(args[1]) == 0 {
		return *filter, nil
	}
	err := json.Unmarshal([]byte(args[1]), filter)
	if err != nil {
		return nil, err
	}
	return *filter, nil
}
