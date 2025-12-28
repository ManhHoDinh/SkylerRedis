package command

import (
	"SkylerRedis/internal/memory"
	"SkylerRedis/internal/utils"
	"net"
)

type Srem struct{}

func (s Srem) Handle(conn net.Conn, args []string, isMaster bool, shard *memory.Shard) {
	if len(args) < 3 {
		utils.WriteError(conn, "ERR wrong number of arguments for 'srem' command")
		return
	}

	key := args[1]
	membersToRemove := args[2:]
	removedCount := 0

	shard.Mu.Lock()
	defer shard.Mu.Unlock()

	set, ok := shard.Sets[key]
	if !ok {
		utils.WriteInteger(conn, 0)
		return
	}

	for _, member := range membersToRemove {
		if _, ok := set[member]; ok {
			delete(set, member)
			removedCount++
		}
	}

	utils.WriteInteger(conn, removedCount)
}
