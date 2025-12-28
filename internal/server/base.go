package server

import (
	"SkylerRedis/internal/entity"
)

type Server interface {
	IsMaster() bool
	HandShake()
}

// NewMaster creates a new Master server instance.
func NewMaster() Server {
	return &Master{
		BaseServer: &entity.BaseServer{
			IsMasterServer: true,
		},
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
	}
}
