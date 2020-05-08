package domain

type ActiveSorts func(names []string) []Sort

type ActiveFilters func(names []string) []Filter

type StateManager interface {
	PutState(key string, value []byte) error
	GetState(key string) ([]byte, error)
}

type QueueContract interface {
	Put(StateManager, ...Item) error
	Peek(StateManager) (*Item, error)
	AddSort(StateManager, ...string) error
	RemoveSort(StateManager, ...string) error
	AddFilter(StateManager, ...string) error
	RemoveFilter(StateManager, ...string) error
	GetState(StateManager) (*QueueContext, error)
	SetState(StateManager, *QueueContext) error
}
