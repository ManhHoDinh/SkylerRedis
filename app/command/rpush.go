package command

import (
	"net"
	"skylerRedis/app/memory"
	"skylerRedis/app/utils"
)

func handleRPush(conn net.Conn, args []string) {
	if len(args) < 3 {
		utils.WriteError(conn, "wrong number of arguments for 'RPUSH'")
		return
	}

	key := args[1]

	memory.Mu.Lock()
	defer memory.Mu.Unlock()

	for i := 2; i < len(args); i++ {
		memory.RPush[key] = append(memory.RPush[key], args[i])
	}

	wakeUpFirstBlocking(key)

	utils.WriteInteger(conn, len(memory.RPush[key]))
}
