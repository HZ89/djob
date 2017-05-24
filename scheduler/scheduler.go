package scheduler

import "errors"

// A map of registered scheduler
var schedulers = make(map[string]Scheduler)
var ErrAlreadyExists = errors.New("Scheduler already registered")

type Scheduler interface {
}

func Register(schedulerName string, scheduler Scheduler) error {
	if _, exists := schedulers[schedulerName]; exists {
		return ErrAlreadyExists
	}
	schedulers[schedulerName] = scheduler
	return nil
}

func ListsAllSchedulers() *[]string {
	keys := make([]string, 0, len(schedulers))
	for k := range schedulers {
		keys = append(keys, k)
	}
	return &keys
}
