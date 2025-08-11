package command

import (
	"SkylerRedis/app/memory"
	"SkylerRedis/app/server"
	"SkylerRedis/app/utils"
)

func handleMULTI(server server.Server, args []string) {
	memory.IsMulti[server.Conn] = true
	utils.WriteSimpleString(server.Conn, "OK")
}
