package memory

import (
	"net"
	"sync"
)

var Store = make(map[string]Entry)
var RPush = make(map[string][]string)
var Queue = make(map[net.Conn][][]string, 0)
var IsMulti = make(map[net.Conn]bool)
var (
	Blockings = make(map[string][]BlockingRequest)
	Mu        = sync.Mutex{}
)
