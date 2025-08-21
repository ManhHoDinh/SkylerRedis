package command

import (
	"SkylerRedis/internal/utils"
	"fmt"
	"net"
)

type Ping struct{}

func (Ping) Handle(Conn net.Conn, args []string, isMaster bool) {
	fmt.Println("Received PING command")
	utils.WriteSimpleString(Conn, "PONG")
}
