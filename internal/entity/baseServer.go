package entity

import "net"

type BaseServer struct {
	Addr           string
	Conn           net.Conn
	IsMasterServer bool
}
