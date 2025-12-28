package command

import (
	"SkylerRedis/internal/memory"
	"SkylerRedis/internal/utils"
	"net"
)

// Cmsquery implements the CMS.QUERY command.
type Cmsquery struct{}

func (cmd Cmsquery) Handle(conn net.Conn, args []string, isMaster bool, shard *memory.Shard) {
	if len(args) != 3 {
		utils.WriteError(conn, "ERR wrong number of arguments for 'cms.query' command")
		return
	}

	key := args[1]
	item := args[2]

	shard.Mu.Lock()
	defer shard.Mu.Unlock()

	sketch, ok := shard.CountMinSketches[key]
	if !ok {
		utils.WriteInteger(conn, 0)
		return
	}

	count := sketch.Count([]byte(item))
	utils.WriteInteger(conn, int(count))
}
