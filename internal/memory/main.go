package memory

import (
	"SkylerRedis/internal/entity"
	"net"
	"sync"
)

var (
	RPush     = make(map[string][]string)
	Store     = make(map[string]Entry)
	Blockings = make(map[string][]BlockingRequest)
	Mu        = sync.Mutex{}
	Queue     = make(map[net.Conn][][]string, 0)
	IsMulti   = make(map[net.Conn]bool)
	Stream    = map[string]StreamEntry{}
	Slaves    = make([]entity.BaseServer, 0)
	OffSet    = 0
	StreamIDs = make([]string, 0)
)
