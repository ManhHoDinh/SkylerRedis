package command

import (
	"SkylerRedis/internal/memory"
	"SkylerRedis/internal/utils"
	"net"
)

type Scard struct{}

func (s Scard) Handle(conn net.Conn, args []string, isMaster bool, shard *memory.Shard) {
	if len(args) != 2 {
		utils.WriteError(conn, "ERR wrong number of arguments for 'scard' command")
		return
	}

	key := args[1]
	
	shard.Mu.Lock()
	defer shard.Mu.Unlock()

	set, ok := shard.Sets[key]
	if !ok {
		utils.WriteInteger(conn, 0)
		return
	}

	utils.WriteInteger(conn, len(set))
}
