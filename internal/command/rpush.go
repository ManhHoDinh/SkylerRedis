package command

import (
	"SkylerRedis/internal/memory"
	"SkylerRedis/internal/utils"
	"net"
)

type RPush struct{}

func (RPush) Handle(Conn net.Conn, args []string, isMaster bool) {
	if len(args) < 3 {
		utils.WriteError(Conn, "wrong number of arguments for 'RPUSH'")
		return
	}

	key := args[1]

	memory.Mu.Lock()
	defer memory.Mu.Unlock()

	for i := 2; i < len(args); i++ {
		memory.RPush[key] = append(memory.RPush[key], args[i])
	}

	utils.WriteInteger(Conn, len(memory.RPush[key]))
	wakeUpFirstBlocking(key)
}
