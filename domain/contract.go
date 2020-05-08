package domain

import (
	"encoding/json"
	"sort"

	"github.com/google/uuid"
)

type contract struct {
	activeFilters ActiveFilters
	activeSorts   ActiveSorts
}

func NewContract(activeFilters ActiveFilters, activeSorts ActiveSorts) QueueContract {
	return &contract{activeFilters, activeSorts}
}

func (q *contract) Put(stub StateManager, items ...Item) error {
	context, err := q.GetState(stub)
	if err != nil {
		return err
	}
	for _, item := range items {
		item.Id = uuid.New().String()
		context.Items = append(context.Items, item)
	}
	err = q.SetState(stub, context)
	if err != nil {
		return err
	}
	return nil
}

func (q *contract) Peek(stub StateManager) (item *Item, err error) {
	context, err := q.GetState(stub)
	if err != nil {
		return
	}
	filters := q.activeFilters(context.Filters)
	currentItems := context.Items
	var items []Item
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
	var resultItems []Item
	for _, v := range currentItems {
		if v.Id == item.Id {
			continue
		}
		resultItems = append(resultItems, v)
	}
	context.Items = resultItems
	err = q.SetState(stub, context)
	if err != nil {
		return nil, err
	}
	return
}

func (q *contract) AddFilter(stub StateManager, filters ...string) error {
	context, err := q.GetState(stub)
	if err != nil {
		return err
	}
	context.Filters = append(context.Filters, filters...)
	err = q.SetState(stub, context)
	if err != nil {
		return err
	}
	return nil
}

func (q *contract) RemoveFilter(stub StateManager, filters ...string) error {
	context, err := q.GetState(stub)
	if err != nil {
		return err
	}
	fMap := make(map[string]struct{})
	for _, f := range filters {
		fMap[f] = struct{}{}
	}
	var remainingF []string
	for _, f := range context.Filters {
		if _, ok := fMap[f]; ok {
			continue
		}
		remainingF = append(remainingF, f)
	}
	context.Filters = remainingF
	err = q.SetState(stub, context)
	if err != nil {
		return err
	}
	return nil
}

func (q *contract) AddSort(stub StateManager, sorts ...string) error {
	context, err := q.GetState(stub)
	if err != nil {
		return err
	}
	context.Sorts = append(context.Sorts, sorts...)
	err = q.SetState(stub, context)
	if err != nil {
		return err
	}
	return nil
}

func (q *contract) RemoveSort(stub StateManager, sorts ...string) error {
	context, err := q.GetState(stub)
	if err != nil {
		return err
	}
	sMap := make(map[string]struct{})
	for _, f := range sorts {
		sMap[f] = struct{}{}
	}
	var remainingS []string
	for _, s := range context.Sorts {
		if _, ok := sMap[s]; ok {
			continue
		}
		remainingS = append(remainingS, s)
	}
	context.Sorts = remainingS
	err = q.SetState(stub, context)
	if err != nil {
		return err
	}
	return nil
}

func (q *contract) GetState(stub StateManager) (model *QueueContext, err error) {
	existing, err := stub.GetState("QueueContext")
	if err != nil {
		return
	}
	model = new(QueueContext)
	err = json.Unmarshal(existing, model)
	if err != nil {
		return nil, err
	}
	return
}

func (q *contract) SetState(stub StateManager, model *QueueContext) error {
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
