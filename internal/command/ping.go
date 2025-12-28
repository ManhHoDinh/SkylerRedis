package command

import (
	"SkylerRedis/internal/memory"
	"SkylerRedis/internal/utils"
	"fmt"
	"net"
)

type Ping struct{}

func (Ping) Handle(Conn net.Conn, args []string, isMaster bool, shard *memory.Shard) {
	fmt.Println("Received PING command")
	utils.WriteSimpleString(Conn, "PONG")
}
