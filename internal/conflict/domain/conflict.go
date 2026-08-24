package domain

import "time"

type Record struct {
	ID, Address, Reason string
	DetectedAt          time.Time
	Resolved            bool
	State               string
}

const (
	StateDetected     = "detected"
	StateCompensating = "compensating"
	StateResolved     = "resolved"
)

func NewRecord(id, address, reason string, detectedAt time.Time) Record {
	return Record{ID: id, Address: address, Reason: reason, DetectedAt: detectedAt, State: StateDetected}
}

func CanTransition(from, to string) bool {
	return from == StateDetected && to == StateCompensating
}
