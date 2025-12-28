package command

import (
	"SkylerRedis/internal/memory"
	"SkylerRedis/internal/utils"
	"net"
)

type Echo struct{}

func (Echo) Handle(Conn net.Conn, args []string, isMaster bool, masterReplID string, masterReplOffset int, connectedSlaves int, shard *memory.Shard) {
	if len(args) != 2 {
		utils.WriteError(Conn, "wrong number of arguments for 'ECHO'")
		return
	}
	utils.WriteBulkString(Conn, args[1])
}
