package command

import (
	"SkylerRedis/internal/memory"
	"SkylerRedis/internal/utils"
	"net"
)

type Sadd struct{}

func (s Sadd) Handle(conn net.Conn, args []string, isMaster bool, shard *memory.Shard) {
	if len(args) < 3 {
		utils.WriteError(conn, "ERR wrong number of arguments for 'sadd' command")
		return
	}

	key := args[1]
	members := args[2:]
	addedCount := 0

	shard.Mu.Lock()
	defer shard.Mu.Unlock()

	// Check if the set exists, if not, create it.
	if _, ok := shard.Sets[key]; !ok {
		shard.Sets[key] = make(map[string]struct{})
	}

	set := shard.Sets[key]

	for _, member := range members {
		// Check if the member already exists in the set.
		if _, ok := set[member]; !ok {
			set[member] = struct{}{}
			addedCount++
		}
	}

	utils.WriteInteger(conn, addedCount)
}
