package command

import (
	"SkylerRedis/app/utils"
	"net"
)

func handleEcho(conn net.Conn, args []string) {
	if len(args) != 2 {
		utils.WriteError(conn, "wrong number of arguments for 'ECHO'")
		return
	}
	utils.WriteSimpleString(conn, args[1])
}
