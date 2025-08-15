package memory

import (
	"SkylerRedis/app/server"
	"net"
	"sync"
)

var Store = make(map[string]Entry)
var RPush = make(map[string][]string)
var Queue = make(map[net.Conn][][]string, 0)
var IsMulti = make(map[net.Conn]bool)
var Stream = map[string]StreamEntry{}
var StreamIDs = make([]string, 0)
var (
	Blockings = make(map[string][]BlockingRequest)
	Mu        = sync.Mutex{}
)
var Master = &server.Master{Slaves: make([]*server.Slave, 0)}
var OffSet int = 0
