package domain

type ActiveSorts func(filters SortRequest) ([]Sort, error)
type ActiveFilters func(sorts FilterRequest) ([]Filter, error)
type Validators func() ([]Validator, error)

type StateManager interface {
	PutState(key string, value []byte) error
	GetState(key string) ([]byte, error)
}

type QueueContract interface {
	Put(StateManager, ...Item) ([]string, error)
	Peek(StateManager, SortRequest, FilterRequest) (*Item, error)
	Update(StateManager, Payload, FilterRequest) error
	Init(StateManager, *QueueContext) error
}
