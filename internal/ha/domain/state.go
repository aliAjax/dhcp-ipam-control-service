package domain

type Mode string

const (
	ActiveActive  Mode = "active-active"
	ActiveStandby Mode = "active-standby"
)

type Event struct {
	LeaseID string
	Version uint64
	State   string
	Node    string
}
