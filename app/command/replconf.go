package command

import (
	"SkylerRedis/app/server"
	"SkylerRedis/app/utils"
	"fmt"
)

func handleREPLCONF(server server.Server, args []string) {
	fmt.Println("Handling REPLCONF command with args:", args)
	utils.WriteSimpleString(server.Conn, "OK")
}