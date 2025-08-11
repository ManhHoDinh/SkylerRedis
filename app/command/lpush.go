package command

import (
	"SkylerRedis/app/memory"
	"SkylerRedis/app/server"
	"SkylerRedis/app/utils"
)

func handleLPush(server server.Server, args []string) {
	if len(args) < 3 {
		utils.WriteError(server.Conn, "wrong number of arguments for 'LPUSH'")
		return
	}

	key := args[1]

	memory.Mu.Lock()
	defer memory.Mu.Unlock()

	for i := 2; i < len(args); i++ {
		memory.RPush[key] = append([]string{args[i]}, memory.RPush[key]...)
	}

	// Wake up blocked BLPOP clients if any
	wakeUpFirstBlocking(key)
	utils.WriteInteger(server.Conn, len(memory.RPush[key]))
}

func wakeUpFirstBlocking(key string) {
	if list, ok := memory.Blockings[key]; ok && len(list) > 0 {
		req := list[0]
		memory.Blockings[key] = list[1:]
		select {
		case req.Ch <- key:
		default:
		}
	}
}
