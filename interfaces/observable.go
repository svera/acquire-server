package interfaces

// Observable is an interface that defines the minimum set of functions needed
// to implement an event observer
type Observable interface {
	On(event string, fn interface{}) Observable
	Trigger(event string, arguments ...interface{}) Observable
}
