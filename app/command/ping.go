package command

import (
	"SkylerRedis/app/server"
	"SkylerRedis/app/utils"
	"fmt"
)

func handlePing(sever server.Server) {
	fmt.Println("Received PING command")
	utils.WriteSimpleString(sever.Conn, "PONG")
}
