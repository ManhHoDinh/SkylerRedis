package server

import (
	"SkylerRedis/internal/entity"
	"SkylerRedis/internal/utils" // Added for ArgsToRESP
	"crypto/rand"
	"encoding/hex"
	"fmt" // Added for Fprintf
	"log"
	"net" // Added for net.Dial and net.Conn
	"sync"
)

type Master struct {
	*entity.BaseServer
	MasterReplID     string
	MasterReplOffset int
	Slaves           []entity.BaseServer // List of connected slaves
	SlavesMu         sync.Mutex          // Mutex to protect access to Slaves
}

// generateReplicationID creates a random 40-character hexadecimal string.
func generateReplicationID() string {
	bytes := make([]byte, 20)
	if _, err := rand.Read(bytes); err != nil {
		log.Fatalf("Failed to generate replication ID: %v", err)
	}
	return hex.EncodeToString(bytes)
}

// HandShake for a master node is a no-op, but it must satisfy the interface.
func (m *Master) HandShake() {
	// A master does not perform a handshake with another master.
}

// IsMaster returns true for the Master server type.
func (m *Master) IsMaster() bool {
	return true
}

// RegisterSlave registers a new slave with the master.
func (m *Master) RegisterSlave(conn net.Conn, port string) error {
	remoteHost, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
	slaveAddr := utils.FormatAddr(remoteHost, port)

	m.SlavesMu.Lock()
	defer m.SlavesMu.Unlock()

	slave := entity.BaseServer{
		Conn: conn,
		Addr: slaveAddr,
	}
	m.Slaves = append(m.Slaves, slave)
	fmt.Printf("Master: Registered slave %s\n", slaveAddr)
	return nil
}

// PropagateCommand sends a command to all connected slaves.
func (m *Master) PropagateCommand(args []string) {
	resp := utils.ArgsToRESP(args)
	respBytes := []byte(resp)

	m.SlavesMu.Lock()
	defer m.SlavesMu.Unlock()

	for i := len(m.Slaves) - 1; i >= 0; i-- { // Iterate backwards to safely remove disconnected slaves
		slave := m.Slaves[i]
		n, err := slave.Conn.Write(respBytes)
		if err != nil {
			fmt.Printf("Error propagating command to slave %s: %v. Removing slave.\n", slave.Addr, err)
			m.Slaves = append(m.Slaves[:i], m.Slaves[i+1:]...) // Remove disconnected slave
			continue
		}
		m.MasterReplOffset += n // Update offset based on bytes written
	}
}
