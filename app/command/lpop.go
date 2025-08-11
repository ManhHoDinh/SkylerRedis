package command

import (
	"SkylerRedis/app/memory"
	"SkylerRedis/app/server"
	"SkylerRedis/app/utils"
	"fmt"
	"strconv"
)

func handleLPop(server server.Server, args []string) {
	if len(args) < 2 {
		utils.WriteError(server.Conn, "wrong number of arguments for 'LPOP'")
		return
	}
	key := args[1]
	list := memory.RPush[key]
	if len(list) == 0 {
		utils.WriteNull(server.Conn)
		return
	}
	if len(args) == 3 {
		count, err := strconv.Atoi(args[2])
		if err != nil || count <= 0 {
			count = 1
		}
		if count > len(list) {
			count = len(list)
		}
		memory.RPush[key] = list[count:]
		server.Conn.Write([]byte(fmt.Sprintf("*%d\r\n", count)))
		for i := 0; i < count; i++ {
			utils.WriteBulkString(server.Conn, list[i])
		}
	} else {
		memory.RPush[key] = list[1:]
		utils.WriteBulkString(server.Conn, list[0])
	}
}
