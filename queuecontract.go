package queuecontract

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/andriiyaremenko/queuecontract/domain"
	"github.com/google/uuid"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

type queueContract struct {
	activeFilters domain.ActiveFilters
	activeSorts   domain.ActiveSorts
}

func NewQueueContract(
	activeFilters domain.ActiveFilters, activeSorts domain.ActiveSorts,
) domain.QueueContract {
	return &queueContract{
		activeFilters: activeFilters,
		activeSorts:   activeSorts,
	}
}

func (q *queueContract) Init(stub shim.ChaincodeStubInterface) pb.Response {
	q.setState(stub, new(domain.QueueContext))
	return shim.Success(nil)
}

func (q *queueContract) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	switch function {
	case "Init":
		return q.Init(stub)
	case "Put":
		if len(args) == 0 {
			return shim.Error("Put: missing arguments")
		}
		if len(args) > 1 {
			return shim.Error("Put: to much arguments")
		}
		raw := args[0]
		item := new(domain.Item)
		err := json.Unmarshal([]byte(raw), item)
		if err != nil {
			return shim.Error(fmt.Sprintf("Put: %v", err))
		}
		err = q.Put(stub, *item)
		if err != nil {
			return shim.Error(fmt.Sprintf("Put: %v", err))
		}
		return shim.Success(nil)
	case "Peek":
		if len(args) > 0 {
			return shim.Error("Peek: to much arguments")
		}
		item, err := q.Peek(stub)
		if err != nil {
			return shim.Error(fmt.Sprintf("Peek: %v", err))
		}
		raw, err := json.Marshal(item)
		if err != nil {
			return shim.Error(fmt.Sprintf("Peek: %v", err))
		}
		return shim.Success(raw)
	case "AddFilter":
		if len(args) == 0 {
			return shim.Error("AddFilter: missing arguments")
		}
		if len(args) > 1 {
			return shim.Error("AddFilter: to much arguments")
		}
		err := q.AddFilter(stub, args[0])
		if err != nil {
			return shim.Error(fmt.Sprintf("AddFilter: %v", err))
		}
		return shim.Success(nil)
	case "RemoveFilter":
		if len(args) == 0 {
			return shim.Error("RemoveFilter: missing arguments")
		}
		if len(args) > 1 {
			return shim.Error("RemoveFilter: to much arguments")
		}
		err := q.RemoveFilter(stub, args[0])
		if err != nil {
			return shim.Error(fmt.Sprintf("RemoveFilter: %v", err))
		}
		return shim.Success(nil)
	case "AddSort":
		if len(args) == 0 {
			return shim.Error("AddSort: missing arguments")
		}
		if len(args) > 1 {
			return shim.Error("AddSort: to much arguments")
		}
		err := q.AddSort(stub, args[0])
		if err != nil {
			return shim.Error(fmt.Sprintf("AddSort: %v", err))
		}
		return shim.Success(nil)
	case "RemoveSort":
		if len(args) == 0 {
			return shim.Error("RemoveSort: missing arguments")
		}
		if len(args) > 1 {
			return shim.Error("RemoveSort: to much arguments")
		}
		err := q.RemoveSort(stub, args[0])
		if err != nil {
			return shim.Error(fmt.Sprintf("RemoveSort: %v", err))
		}
		return shim.Success(nil)
	default:
		return shim.Error("invalid invoke function name.")
	}
}

func (q *queueContract) Put(stub shim.ChaincodeStubInterface, item domain.Item) error {
	context, err := q.getState(stub)
	if err != nil {
		return err
	}
	item.Id = uuid.New().String()
	context.Items = append(context.Items, item)
	err = q.setState(stub, context)
	if err != nil {
		return err
	}
	return nil
}

func (q *queueContract) Peek(stub shim.ChaincodeStubInterface) (item *domain.Item, err error) {
	context, err := q.getState(stub)
	if err != nil {
		return
	}
	filters := q.activeFilters(context.Filters)
	currentItems := context.Items
	var items []domain.Item
filterLoop:
	for _, item := range currentItems {
		for _, filter := range filters {
			if !filter.F(item) {
				continue filterLoop
			}
		}
		items = append(items, item)
	}
	sorts := q.activeSorts(context.Sorts)
	for _, s := range sorts {
		sort.Slice(items, func(first, second int) bool {
			return s.F(items[first], items[second])
		})
	}
	if len(items) == 0 {
		return nil, nil
	}
	item = &items[0]
	var resultItems []domain.Item
	for _, v := range currentItems {
		if v.Id == item.Id {
			continue
		}
		resultItems = append(resultItems, v)
	}
	context.Items = resultItems
	err = q.setState(stub, context)
	if err != nil {
		return nil, err
	}
	return
}

func (q *queueContract) AddFilter(stub shim.ChaincodeStubInterface, filter string) error {
	context, err := q.getState(stub)
	if err != nil {
		return err
	}
	context.Filters = append(context.Filters, filter)
	err = q.setState(stub, context)
	if err != nil {
		return err
	}
	return nil
}

func (q *queueContract) RemoveFilter(stub shim.ChaincodeStubInterface, filter string) error {
	context, err := q.getState(stub)
	if err != nil {
		return err
	}
	var filters []string
	for _, f := range context.Filters {
		if filter == f {
			continue
		}
		filters = append(filters, f)
	}
	context.Filters = filters
	err = q.setState(stub, context)
	if err != nil {
		return err
	}
	return nil
}

func (q *queueContract) AddSort(stub shim.ChaincodeStubInterface, sort string) error {
	context, err := q.getState(stub)
	if err != nil {
		return err
	}
	context.Sorts = append(context.Sorts, sort)
	err = q.setState(stub, context)
	if err != nil {
		return err
	}
	return nil
}

func (q *queueContract) RemoveSort(stub shim.ChaincodeStubInterface, sort string) error {
	context, err := q.getState(stub)
	if err != nil {
		return err
	}
	var sorts []string
	for _, s := range context.Sorts {
		if sort == s {
			continue
		}
		sorts = append(sorts, s)
	}
	context.Sorts = sorts
	err = q.setState(stub, context)
	if err != nil {
		return err
	}
	return nil
}

func (q *queueContract) getState(stub shim.ChaincodeStubInterface) (model *domain.QueueContext, err error) {
	existing, err := stub.GetState("QueueContext")
	if err != nil {
		return
	}
	model = new(domain.QueueContext)
	err = json.Unmarshal(existing, model)
	if err != nil {
		return nil, err
	}
	return
}

func (q *queueContract) setState(stub shim.ChaincodeStubInterface, model *domain.QueueContext) error {
	state, err := json.Marshal(model)
	if err != nil {
		return err
	}
	err = stub.PutState("QueueContext", state)
	if err != nil {
		return err
	}
	return nil
}
