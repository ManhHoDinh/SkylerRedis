package command

import (
	"SkylerRedis/internal/memory"
	"SkylerRedis/internal/utils"
	"fmt"
	"net"
	"strconv"
)

type LPop struct{}

func (LPop) Handle(Conn net.Conn, args []string, isMaster bool, masterReplID string, masterReplOffset int, connectedSlaves int, shard *memory.Shard) {
	if len(args) < 2 {
		utils.WriteError(Conn, "wrong number of arguments for 'LPOP'")
		return
	}
	key := args[1]

	shard.Mu.Lock()
	defer shard.Mu.Unlock()

	list := shard.RPush[key]
	if len(list) == 0 {
		utils.WriteNull(Conn)
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
		shard.RPush[key] = list[count:]
		Conn.Write([]byte(fmt.Sprintf("*%d\r\n", count)))
		for i := 0; i < count; i++ {
			utils.WriteBulkString(Conn, list[i])
		}
	} else {
		shard.RPush[key] = list[1:]
		utils.WriteBulkString(Conn, list[0])
	}
}
