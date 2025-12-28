package server

import "SkylerRedis/internal/entity"

type Master struct {
	*entity.BaseServer
}

// HandShake for a master node is a no-op, but it must satisfy the interface.
func (m *Master) HandShake() {
	// A master does not perform a handshake with another master.
}

// IsMaster returns true for the Master server type.
func (m *Master) IsMaster() bool {
	return true
}
