package command

import (
	"SkylerRedis/internal/memory"
	"SkylerRedis/internal/utils"
	"net"
)

type LPush struct{}

func (LPush) Handle(Conn net.Conn, args []string, isMaster bool, shard *memory.Shard) {
	if len(args) < 3 {
		utils.WriteError(Conn, "wrong number of arguments for 'LPUSH'")
		return
	}

	key := args[1]

	shard.Mu.Lock()
	defer shard.Mu.Unlock()

	for i := 2; i < len(args); i++ {
		shard.RPush[key] = append([]string{args[i]}, shard.RPush[key]...)
	}

	// Wake up blocked BLPOP clients if any
	utils.WriteInteger(Conn, len(shard.RPush[key]))
	// The wakeUpFirstBlocking function needs to be called after shard.Mu.Unlock()
	// to avoid deadlock, but also needs to access the global memory.Blockings safely.
	// For now, it remains global.
	wakeUpFirstBlocking(key) // This function still accesses global memory.Blockings, which uses its own memory.Mu.
}
