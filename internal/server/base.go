package server

import (
	"SkylerRedis/internal/entity"
	"net" // Added for RegisterSlave method
	"sync" // Added for sync.Mutex
)

type Server interface {
	IsMaster() bool
	HandShake()
	RegisterSlave(conn net.Conn, port string) error // New method for master to register a slave
}

// NewMaster creates a new Master server instance.
func NewMaster() Server {
	return &Master{
		BaseServer: &entity.BaseServer{
			IsMasterServer: true,
		},
		MasterReplID:     generateReplicationID(),
		MasterReplOffset: 0,
		Slaves:           make([]entity.BaseServer, 0),
		SlavesMu:         sync.Mutex{},
	}
}
// NewSlave creates a new Slave server instance.
func NewSlave(replicaof *string, port *string) Server {
	return &Slave{
		BaseServer: &entity.BaseServer{
			IsMasterServer: false,
		},
		ReplicaOf: replicaof,
		Port:      port,
		Offset:    0,
	}
}
