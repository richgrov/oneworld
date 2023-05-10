package oneworld

type Entity interface {
	Id() int32
	Tick()
}
