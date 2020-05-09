package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"
)

type contract struct {
	activeFilters ActiveFilters
	activeSorts   ActiveSorts
	validators    Validators
}

func NewContract(activeFilters ActiveFilters, activeSorts ActiveSorts, validators Validators) QueueContract {
	return &contract{activeFilters, activeSorts, validators}
}

func (q *contract) Init(stub StateManager, model *QueueContext) error {
	return q.setState(stub, model)
}

func (q *contract) Put(stub StateManager, items ...Item) (ids []string, err error) {
	validators, err := q.validators()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("QueueContract.Put: %v", err))
	}
	var sb strings.Builder
	sb.WriteString("QueueContract.Put: [\n")
	hasErrors := false
	for _, v := range validators {
		for _, i := range items {
			if err := v.F(i.Data); err != nil {
				hasErrors = true
				sb.WriteString(fmt.Sprintf("\tFailed validation: %s: %v\n", v.Name, err))
			}
		}
	}
	if hasErrors {
		sb.WriteString("]")
		return nil, errors.New(sb.String())
	}
	context, err := q.getState(stub)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("QueueContract.Put: %v", err))
	}
	for _, item := range items {
		id := uuid.New().String()
		ids = append(ids, id)
		item.Id = id
		context.Items = append(context.Items, item)
	}
	err = q.setState(stub, context)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("QueueContract.Put: %v", err))
	}
	return
}

func (q *contract) Peek(stub StateManager, sorts SortRequest, filters FilterRequest) (item *Item, err error) {
	context, err := q.getState(stub)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("QueueContract.Peek: %v", err))
	}
	fs, err := q.activeFilters(filters)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("QueueContract.Peek: %v", err))
	}
	currentItems := context.Items
	var items []Item
filterLoop:
	for _, item := range currentItems {
		for _, filter := range fs {
			if !filter.F(item, filters[filter.Name]...) {
				continue filterLoop
			}
		}
		items = append(items, item)
	}
	ss, err := q.activeSorts(sorts)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("QueueContract.Peek: %v", err))
	}
	for _, s := range ss {
		sort.Slice(items, func(first, second int) bool {
			return s.F(items[first], items[second], sorts[s.Name]...)
		})
	}
	if len(items) == 0 {
		return nil, nil
	}
	item = &items[0]
	var resultItems []Item
	for _, v := range currentItems {
		if v.Id == item.Id {
			continue
		}
		resultItems = append(resultItems, v)
	}
	context.Items = resultItems
	err = q.setState(stub, context)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("QueueContract.Peek: %v", err))
	}
	return
}

func (q *contract) Update(stub StateManager, data Payload, filters FilterRequest) error {
	context, err := q.getState(stub)
	if err != nil {
		return errors.New(fmt.Sprintf("QueueContract.Peek: %v", err))
	}
	fs, err := q.activeFilters(filters)
	if err != nil {
		return errors.New(fmt.Sprintf("QueueContract.Peek: %v", err))
	}
	currentItems := context.Items
	var items []Item
filterLoop:
	for _, item := range currentItems {
		for _, filter := range fs {
			if !filter.F(item) {
				continue filterLoop
			}
		}
		item.Data = data
		items = append(items, item)
	}
	return nil
}

func (q *contract) getState(stub StateManager) (model *QueueContext, err error) {
	existing, err := stub.GetState("QueueContext")
	if err != nil {
		return nil, errors.New(fmt.Sprintf("QueueContract.getState: %v", err))
	}
	model = new(QueueContext)
	err = json.Unmarshal(existing, model)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("QueueContract.getState: %v", err))
	}
	return
}

func (q *contract) setState(stub StateManager, model *QueueContext) error {
	state, err := json.Marshal(model)
	if err != nil {
		return errors.New(fmt.Sprintf("QueueContract.setState: %v", err))
	}
	err = stub.PutState("QueueContext", state)
	if err != nil {
		return errors.New(fmt.Sprintf("QueueContract.setState: %v", err))
	}
	return nil
}
