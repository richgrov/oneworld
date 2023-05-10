package oneworld

type PlayerJoinEvent struct {
	player *Player
}

func (e *PlayerJoinEvent) Player() *Player {
	return e.player
}
