package command

import (
	"SkylerRedis/internal/memory"
	"SkylerRedis/internal/utils"
	"net"
)

type Smembers struct{}

func (s Smembers) Handle(conn net.Conn, args []string, isMaster bool, masterReplID string, masterReplOffset int, connectedSlaves int, shard *memory.Shard) {
	if len(args) != 2 {
		utils.WriteError(conn, "ERR wrong number of arguments for 'smembers' command")
		return
	}

	key := args[1]

	shard.Mu.Lock()
	defer shard.Mu.Unlock()

	set, ok := shard.Sets[key]
	if !ok {
		utils.WriteArray(conn, []string{})
		return
	}

	members := make([]string, 0, len(set))
	for member := range set {
		members = append(members, member)
	}

	utils.WriteArray(conn, members)
}
