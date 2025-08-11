package server

import (
	"net"
)

type Server struct {
	SeverId     int
	Addr        string
	IsMaster    bool
	IsConnected bool
	Conn       net.Conn
}
