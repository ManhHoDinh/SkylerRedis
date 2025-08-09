package command

import (
	"SkylerRedis/app/utils"
	"fmt"
	"net"
)

func handlePing(conn net.Conn) {
	fmt.Println("Received PING command")
	utils.WriteSimpleString(conn, "PONG")
}
