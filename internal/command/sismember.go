package command

import (
	"SkylerRedis/internal/memory"
	"SkylerRedis/internal/utils"
	"net"
)

type Sismember struct{}

func (s Sismember) Handle(conn net.Conn, args []string, isMaster bool, masterReplID string, masterReplOffset int, connectedSlaves int, shard *memory.Shard) {
	if len(args) != 3 {
		utils.WriteError(conn, "ERR wrong number of arguments for 'sismember' command")
		return
	}

	key := args[1]
	member := args[2]

	shard.Mu.Lock()
	defer shard.Mu.Unlock()

	set, ok := shard.Sets[key]
	if !ok {
		utils.WriteInteger(conn, 0)
		return
	}

	if _, ok := set[member]; ok {
		utils.WriteInteger(conn, 1)
	} else {
		utils.WriteInteger(conn, 0)
	}
}
