package command

import (
	"SkylerRedis/app/utils"
	"net"
)

func handlePing(conn net.Conn) {
	utils.WriteSimpleString(conn, "PONG")
}
