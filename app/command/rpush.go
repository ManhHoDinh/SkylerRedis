package command

import (
	"SkylerRedis/app/memory"
	"SkylerRedis/app/server"
	"SkylerRedis/app/utils"
)

func handleRPush(server server.Server, args []string) {
	if len(args) < 3 {
		utils.WriteError(server.Conn, "wrong number of arguments for 'RPUSH'")
		return
	}

	key := args[1]

	memory.Mu.Lock()
	defer memory.Mu.Unlock()

	for i := 2; i < len(args); i++ {
		memory.RPush[key] = append(memory.RPush[key], args[i])
	}

	if wakeUpFirstBlocking(key) {
		utils.WriteInteger(server.Conn, len(memory.RPush[key])-1)
	} else {
		utils.WriteInteger(server.Conn, len(memory.RPush[key]))
	}

}
