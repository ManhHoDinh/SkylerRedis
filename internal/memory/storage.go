package memory

import "time"

type Entry struct {
	Value      string
	ExpiryTime time.Time
}
