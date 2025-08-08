package command

import (
	"net"
	"skylerRedis/app/utils"
)

func handleEcho(conn net.Conn, args []string) {
	if len(args) != 2 {
		utils.WriteError(conn, "wrong number of arguments for 'ECHO'")
		return
	}
	utils.WriteSimpleString(conn, args[1])
}
