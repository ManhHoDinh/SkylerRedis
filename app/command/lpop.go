package command

import (
	"fmt"
	"net"
	"skylerRedis/app/memory"
	"skylerRedis/app/utils"
	"strconv"
)

func handleLPop(conn net.Conn, args []string) {
	if len(args) < 2 {
		utils.WriteError(conn, "wrong number of arguments for 'LPOP'")
		return
	}
	key := args[1]
	list := memory.RPush[key]
	if len(list) == 0 {
		utils.WriteNull(conn)
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
		conn.Write([]byte(fmt.Sprintf("*%d\r\n", count)))
		for i := 0; i < count; i++ {
			utils.WriteBulkString(conn, list[i])
		}
	} else {
		memory.RPush[key] = list[1:]
		utils.WriteBulkString(conn, list[0])
	}
}
