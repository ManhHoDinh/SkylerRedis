package memory

import "time"

type BlockingRequest struct {
	Key     string
	Ch      chan string
	Timeout time.Duration
}
