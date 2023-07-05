package oneworld

type PlayerJoinEvent struct {
	Player *Player
}

type ChatEvent struct {
	Player  *Player
	Message string
}

type CommandEvent struct {
	Player  *Player
	Command string
}
