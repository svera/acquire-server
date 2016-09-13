package observableWrapper

import (
	observable "github.com/GianlucaGuarini/go-observable"
	"github.com/svera/sackson-server/interfaces"
)

// Wrapper is a struct that contains just an instance of go-observable,
// acting as a wrapper so it can adhere to the Observable interface
type Wrapper struct {
	observer *observable.Observable
}

// New returns a new instance of Wrapper
func New() *Wrapper {
	return &Wrapper{
		observer: observable.New(),
	}
}

// On associates a callback to an specific event
func (o *Wrapper) On(event string, fn interface{}) interfaces.Observable {
	o.observer.On(event, fn)
	return o
}

// Trigger runs all the callbacks associated with the specified event
func (o *Wrapper) Trigger(event string, arguments ...interface{}) interfaces.Observable {
	if len(arguments) > 0 {
		o.observer.Trigger(event, arguments)
	} else {
		o.observer.Trigger(event)
	}
	return o
}
