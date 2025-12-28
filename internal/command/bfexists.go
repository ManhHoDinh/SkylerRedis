package command

import (
	"SkylerRedis/internal/memory"
	"SkylerRedis/internal/utils"
	"net"
)

// Bfexists implements the BF.EXISTS command.
type Bfexists struct{}

func (cmd Bfexists) Handle(conn net.Conn, args []string, isMaster bool, shard *memory.Shard) {
	if len(args) != 3 {
		utils.WriteError(conn, "ERR wrong number of arguments for 'bf.exists' command")
		return
	}

	key := args[1]
	item := args[2]

	shard.Mu.Lock()
	defer shard.Mu.Unlock()

	filter, ok := shard.BloomFilters[key]
	if !ok {
		utils.WriteInteger(conn, 0)
		return
	}

	if filter.TestString(item) {
		utils.WriteInteger(conn, 1) // Item may be in the filter
	} else {
		utils.WriteInteger(conn, 0) // Item is definitely not in the filter
	}
}
