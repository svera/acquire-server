package interfaces

// Observer is an interface that defines the minimum set of functions needed
// to implement an observer.
type Observer interface {
	On(evType interface{}, fn func(ev interface{}))
	Trigger(ev interface{})
}
