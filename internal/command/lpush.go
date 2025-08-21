package command

import (
	"SkylerRedis/internal/memory"
	"SkylerRedis/internal/utils"
	"fmt"
	"net"
)

type LPush struct{}

func (LPush) Handle(Conn net.Conn, args []string, isMaster bool) {
	if len(args) < 3 {
		utils.WriteError(Conn, "wrong number of arguments for 'LPUSH'")
		return
	}

	key := args[1]

	memory.Mu.Lock()
	defer memory.Mu.Unlock()

	for i := 2; i < len(args); i++ {
		memory.RPush[key] = append([]string{args[i]}, memory.RPush[key]...)
	}

	// Wake up blocked BLPOP clients if any
	utils.WriteInteger(Conn, len(memory.RPush[key]))
	wakeUpFirstBlocking(key)
}

func wakeUpFirstBlocking(key string) {
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
