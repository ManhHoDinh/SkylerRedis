package command

import (
	"SkylerRedis/app/memory"
	"SkylerRedis/app/server"
	"SkylerRedis/app/utils"
	"fmt"
)

func handlePing(sever server.Server) {
	fmt.Println("Received PING command")
	memory.OffSet += 14
	utils.WriteSimpleString(sever.Conn, "PONG")
}
