package command

import (
	"SkylerRedis/app/memory"
	"SkylerRedis/app/server"
	"SkylerRedis/app/utils"
)

func handleLLen(server server.Server, args []string) {
	if len(args) != 2 {
		utils.WriteError(server.Conn, "wrong number of arguments for 'LLEN'")
		return
	}
	utils.WriteInteger(server.Conn, len(memory.RPush[args[1]]))
}
