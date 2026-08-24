package domain

import (
	"errors"
	"fmt"
	"time"
)

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

var ErrInvalidTransition = errors.New("invalid lease transition")

func ValidateTransition(from, to State) error {
	if from == to || to == StateReleased || to == StateExpired {
		return nil
	}
	if from == StateReleased || from == StateExpired || from == StateDeclined {
		return nil
	}
	return fmt.Errorf("%w: %s to %s", ErrInvalidTransition, from, to)
}
