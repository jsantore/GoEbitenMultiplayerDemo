package SharedData

const (
	WINDOW_WIDTH  = 1500
	WINDOW_HEIGHT = 1000
	PLAYER_WIDTH  = 64
	PLAYER_HEIGHT = 64
	TRAVEL_SPEED  = 2
)

type MPgame struct {
	Players []*Player
	Gold    []*GoldPile
}

type Player struct {
	PlayerID string
	Xloc     int
	Yloc     int
	Score    int
}

type GoldPile struct {
	Xloc int
	Yloc int
}

const (
	STILL = iota
	LEFT
	RIGHT
	UP
	DOWN
)
