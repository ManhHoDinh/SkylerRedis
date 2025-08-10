package command

import (
	"SkylerRedis/app/utils"
	"fmt"
	"net"
)

func handleREPLCONF(conn net.Conn, args []string) {
	fmt.Println("Handling REPLCONF command with args:", args)
	utils.WriteSimpleString(conn, "OK")
}