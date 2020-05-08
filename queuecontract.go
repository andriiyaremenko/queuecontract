package queuecontract

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/andriiyaremenko/queuecontract/domain"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

type queueContract struct {
	c domain.QueueContract
}

func NewQueueChaincode(
	activeFilters domain.ActiveFilters, activeSorts domain.ActiveSorts,
) shim.Chaincode {
	return &queueContract{c: domain.NewContract(activeFilters, activeSorts)}
}

func (q *queueContract) Init(stub shim.ChaincodeStubInterface) pb.Response {
	q.c.SetState(stub, new(domain.QueueContext))
	return shim.Success(nil)
}

func (q *queueContract) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	switch function {
	case "Init":
		return q.Init(stub)
	case "Put":
		if len(args) == 0 {
			return shim.Error("NewQueueChaincode(...).Invoke: Put: missing arguments")
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
		err := q.c.Put(stub, items...)
		if err != nil {
			return shim.Error(fmt.Sprintf("NewQueueChaincode(...).Invoke: Put: %v", err))
		}
		return shim.Success(nil)
	case "Peek":
		if len(args) > 0 {
			return shim.Error("NewQueueChaincode(...).Invoke: Peek: to much arguments")
		}
		item, err := q.c.Peek(stub)
		if err != nil {
			return shim.Error(fmt.Sprintf("NewQueueChaincode(...).Invoke: Peek: %v", err))
		}
		raw, err := json.Marshal(item)
		if err != nil {
			return shim.Error(fmt.Sprintf("NewQueueChaincode(...).Invoke: Peek: %v", err))
		}
		return shim.Success(raw)
	case "AddFilter":
		if len(args) == 0 {
			return shim.Error("NewQueueChaincode(...).Invoke: AddFilter: missing arguments")
		}
		err := q.c.AddFilter(stub, args...)
		if err != nil {
			return shim.Error(fmt.Sprintf("NewQueueChaincode(...).Invoke: AddFilter: %v", err))
		}
		return shim.Success(nil)
	case "RemoveFilter":
		if len(args) == 0 {
			return shim.Error("NewQueueChaincode(...).Invoke: RemoveFilter: missing arguments")
		}
		err := q.c.RemoveFilter(stub, args...)
		if err != nil {
			return shim.Error(fmt.Sprintf("NewQueueChaincode(...).Invoke: RemoveFilter: %v", err))
		}
		return shim.Success(nil)
	case "AddSort":
		if len(args) == 0 {
			return shim.Error("NewQueueChaincode(...).Invoke: AddSort: missing arguments")
		}
		err := q.c.AddSort(stub, args...)
		if err != nil {
			return shim.Error(fmt.Sprintf("NewQueueChaincode(...).Invoke: AddSort: %v", err))
		}
		return shim.Success(nil)
	case "RemoveSort":
		if len(args) == 0 {
			return shim.Error("NewQueueChaincode(...).Invoke: RemoveSort: missing arguments")
		}
		err := q.c.RemoveSort(stub, args...)
		if err != nil {
			return shim.Error(fmt.Sprintf("NewQueueChaincode(...).Invoke: RemoveSort: %v", err))
		}
		return shim.Success(nil)
	default:
		return shim.Error("NewQueueChaincode(...).Invoke: invalid invoke function name.")
	}
}
