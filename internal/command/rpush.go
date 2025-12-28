package command

import (
	"SkylerRedis/internal/memory"
	"SkylerRedis/internal/utils"
	"fmt"
	"net"
)

type RPush struct{}

func (RPush) Handle(Conn net.Conn, args []string, isMaster bool, masterReplID string, masterReplOffset int, connectedSlaves int, shard *memory.Shard) {
	if len(args) < 3 {
		utils.WriteError(Conn, "wrong number of arguments for 'RPUSH'")
		return
	}

	key := args[1]

	shard.Mu.Lock()
	defer shard.Mu.Unlock()

	for i := 2; i < len(args); i++ {
		shard.RPush[key] = append(shard.RPush[key], args[i])
	}

	utils.WriteInteger(Conn, len(shard.RPush[key]))
	wakeUpFirstBlocking(key)
}

// wakeUpFirstBlocking still accesses global memory.Blockings
func wakeUpFirstBlocking(key string) {
	// Acquire the global mutex for memory.Blockings
	memory.Mu.Lock()
	defer memory.Mu.Unlock()

	if list, ok := memory.Blockings[key]; ok && len(list) > 0 {
		req := list[0]
		fmt.Println("Waking up blocking request for key:", key)
		memory.Blockings[key] = list[1:]
		select {
		case req.Ch <- key:
		default:
		}
	}
}
