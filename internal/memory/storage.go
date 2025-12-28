package memory

import "time"

type Entry struct {
	Value      string
	ExpiryTime time.Time
	LRU        uint64 // Represents approximated last access time for LRU eviction
}
