package command

import (
	"SkylerRedis/internal/memory"
	"SkylerRedis/internal/utils"
	"net"
	"strconv"
)

// Cmsincrby implements the CMS.INCRBY command.
type Cmsincrby struct{}

func (cmd Cmsincrby) Handle(conn net.Conn, args []string, isMaster bool, masterReplID string, masterReplOffset int, connectedSlaves int, shard *memory.Shard) {
	if len(args) != 4 {
		utils.WriteError(conn, "ERR wrong number of arguments for 'cms.incrby' command")
		return
	}

	key := args[1]
	item := args[2]

	increment, err := strconv.ParseUint(args[3], 10, 64)
	if err != nil {
		utils.WriteError(conn, "ERR increment value must be a positive integer")
		return
	}

	shard.Mu.Lock()
	defer shard.Mu.Unlock()

	sketch, ok := shard.CountMinSketches[key]
	if !ok {
		// Create a new sketch with default parameters.
		sketch = memory.NewCountMinSketch(5, 100000) // depth, width
		shard.CountMinSketches[key] = sketch
	}

	sketch.Add([]byte(item), increment)

	utils.WriteSimpleString(conn, "OK")
}
