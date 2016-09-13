package interfaces

type Observable interface {
	On(event string, fn interface{}) Observable
	Trigger(event string, arguments ...interface{}) Observable
}
