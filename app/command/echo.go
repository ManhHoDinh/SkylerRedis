package command

import (
	"SkylerRedis/app/server"
	"SkylerRedis/app/utils"
)

func handleEcho(server server.Server, args []string) {
	if len(args) != 2 {
		utils.WriteError(server.Conn, "wrong number of arguments for 'ECHO'")
		return
	}
	utils.WriteSimpleString(server.Conn, args[1])
}
