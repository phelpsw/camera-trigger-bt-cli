package boards

type Board interface {
	Init(name string, debug bool) error

	SetUpdateCallback(func(interface{}) error)
}
