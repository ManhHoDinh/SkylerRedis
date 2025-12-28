package command

import (
	"SkylerRedis/internal/memory"
	"SkylerRedis/internal/utils"
	"net"
)

type LLen struct{}

func (LLen) Handle(Conn net.Conn, args []string, isMaster bool, masterReplID string, masterReplOffset int, connectedSlaves int, shard *memory.Shard) {
	if len(args) != 2 {
		utils.WriteError(Conn, "wrong number of arguments for 'LLEN'")
		return
	}
	shard.Mu.Lock()
	defer shard.Mu.Unlock()
	utils.WriteInteger(Conn, len(shard.RPush[args[1]]))
}
