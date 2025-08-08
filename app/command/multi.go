package command

import (
	"SkylerRedis/app/memory"
	"SkylerRedis/app/utils"
	"net"
)

func handleMULTI(conn net.Conn, args []string) {
	memory.IsMulti[conn] = true
	utils.WriteSimpleString(conn, "OK")
}
