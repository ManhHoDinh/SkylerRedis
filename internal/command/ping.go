package command

import (
	"SkylerRedis/internal/memory"
	"SkylerRedis/internal/utils"
	"fmt"
	"net"
)

type Ping struct{}

func (Ping) Handle(Conn net.Conn, args []string, isMaster bool, masterReplID string, masterReplOffset int, connectedSlaves int, shard *memory.Shard) {
	fmt.Println("Received PING command with args:", args)
	if len(args) > 2 {
		utils.WriteError(Conn, "wrong number of arguments for 'ping' command")
		return
	}

	if len(args) == 2 {
		utils.WriteSimpleString(Conn, args[1])
	} else {
		utils.WriteSimpleString(Conn, "PONG")
	}
}
