package server

import (
	"net"
)

type Server interface {
	HandleConnection(Conn net.Conn)
	HandShake(replicaof *string, port *string)
}
