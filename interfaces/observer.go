package interfaces

// Observer is an interface that defines the minimum set of functions needed
// to implement an observer.
type Observer interface {
	On(name string, fn func(...interface{}))
	Trigger(name string, params ...interface{})
}
