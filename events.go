package oneworld

type PlayerJoinEvent struct {
	player *Player
}

func (e *PlayerJoinEvent) Player() *Player {
	return e.player
}

type ChatEvent struct {
	player  *Player
	Message string
}

func (e *ChatEvent) Player() *Player {
	return e.player
}

type CommandEvent struct {
	player  *Player
	Command string
}

func (e *CommandEvent) Player() *Player {
	return e.player
}
