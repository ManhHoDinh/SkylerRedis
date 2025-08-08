package command

import (
	"net"
	"skylerRedis/app/memory"
	"skylerRedis/app/utils"
)

func handleLLen(conn net.Conn, args []string) {
	if len(args) != 2 {
		utils.WriteError(conn, "wrong number of arguments for 'LLEN'")
		return
	}
	utils.WriteInteger(conn, len(memory.RPush[args[1]]))
}
