package command

import (
	"SkylerRedis/app/memory"
	"SkylerRedis/app/server"
	"SkylerRedis/app/utils"
)

func handleType(server server.Server, args []string) {
	if len(args) != 2 {
		utils.WriteError(server.Conn, "wrong number of arguments for 'TYPE'")
		return
	}
	key := args[1]
	if _, exists := memory.RPush[key]; exists {
		utils.WriteSimpleString(server.Conn, "list")
	} else if _, exists := memory.Store[key]; exists {
		utils.WriteSimpleString(server.Conn, "string")
	} else if _, exists := memory.Stream[key]; exists {
		utils.WriteSimpleString(server.Conn, "stream")
	} else {
		utils.WriteSimpleString(server.Conn, "none")
	}
}