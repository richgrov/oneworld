package oneworld

type EntityBase struct {
	id int32

	x float64
	y float64
	z float64
}

func (entity *EntityBase) Id() int32 {
	return entity.id
}

func (entity *EntityBase) Pos() (float64, float64, float64) {
	return entity.x, entity.y, entity.z
}

func (*EntityBase) OnSpawned() {}
func (*EntityBase) Tick()      {}

type Entity interface {
	Id() int32
	Pos() (float64, float64, float64)
	OnSpawned()
	Tick()
}
