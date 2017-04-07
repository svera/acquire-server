package interfaces

type Observer interface {
	On(name string, fn func(...interface{}))
	Trigger(name string, params ...interface{})
}
