package command

import (
	"SkylerRedis/internal/memory"
	"SkylerRedis/internal/utils"
	"net"
)

type Type struct{}

func (Type) Handle(Conn net.Conn, args []string, isMaster bool, masterReplID string, masterReplOffset int, connectedSlaves int, shard *memory.Shard) {
	if len(args) != 2 {
		utils.WriteError(Conn, "wrong number of arguments for 'TYPE'")
		return
	}
	key := args[1]

	shard.Mu.Lock()
	defer shard.Mu.Unlock()

	if _, exists := shard.RPush[key]; exists {
		utils.WriteSimpleString(Conn, "list")
	} else if _, exists := shard.Store[key]; exists {
		utils.WriteSimpleString(Conn, "string")
	} else if _, exists := shard.Stream[key]; exists {
		utils.WriteSimpleString(Conn, "stream")
	} else {
		utils.WriteSimpleString(Conn, "none")
	}
}
