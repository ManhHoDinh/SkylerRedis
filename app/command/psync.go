package command

import (
	"SkylerRedis/app/utils"
	"net"
)

func handlePSYNC(conn net.Conn, args []string) {
	utils.WriteSimpleString(conn, "FULLRESYNC 8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb 0")
}
