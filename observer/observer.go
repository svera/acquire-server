// Package observer implements a simple observer struct and methods.
package observer

import (
	"fmt"
)

// Observer holds the functions to be called when registered events are triggered.
type Observer struct {
	callbacks map[string][]func(...interface{})
}

// New returns a new Observer instance.
func New() *Observer {
	return &Observer{
		callbacks: make(map[string][]func(...interface{})),
	}
}

// On registers a function that will be called when the specified event is triggered.
func (o *Observer) On(name string, fn func(...interface{})) {
	o.callbacks[name] = append(o.callbacks[name], fn)
}

// Trigger executes all functions associated to the specified event,
// in the same order they were added.
func (o *Observer) Trigger(name string, params ...interface{}) {
	if callbacks, ok := o.callbacks[name]; ok {
		for _, callback := range callbacks {
			callback(params...)
		}
		return
	}
	panic(fmt.Sprintf("Event %s does not exist", name))
}
