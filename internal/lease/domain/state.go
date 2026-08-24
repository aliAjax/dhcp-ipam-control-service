package domain

import "time"

type State string

const (
	StateOffered  State = "offered"
	StateActive   State = "active"
	StateReleased State = "released"
	StateExpired  State = "expired"
	StateDeclined State = "declined"
)

type Timeline struct {
	State  State
	At     time.Time
	Reason string
}

func (t Timeline) Terminal() bool {
	return t.State == StateReleased || t.State == StateExpired || t.State == StateDeclined
}
