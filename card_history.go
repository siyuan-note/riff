package riff

import "time"

type BaseHistory struct {
	HID          string
	CID          string
	Update       time.Time
	UpdateResult int
	State        State
	Tag          string
	Flag         string
	Suspend      bool
	Priority     float64
	Due          time.Time
	NDues        map[Rating]time.Time
	algo         string
	AlgoImpl     interface{}
}

type ReviewLog struct {
	HID    string
	Rate   Rating
	Review time.Time
}
