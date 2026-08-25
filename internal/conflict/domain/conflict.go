package domain

import "time"

type Record struct {
	ID, Address, Reason string
	DetectedAt          time.Time
	Resolved            bool
}
