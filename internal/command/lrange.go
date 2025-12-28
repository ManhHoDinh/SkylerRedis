package command

import (
	"SkylerRedis/internal/memory"
	"SkylerRedis/internal/utils"
	"fmt"
	"net"
	"strconv"
)

type LRange struct{}

func (LRange) Handle(Conn net.Conn, args []string, isMaster bool, masterReplID string, masterReplOffset int, connectedSlaves int, shard *memory.Shard) {
	if len(args) != 4 {
		utils.WriteError(Conn, "wrong number of arguments for 'LRANGE'")
		return
	}
	key := args[1]
	start, err1 := strconv.Atoi(args[2])
	end, err2 := strconv.Atoi(args[3])
	if err1 != nil || err2 != nil {
		utils.WriteError(Conn, "invalid start or end index")
		return
	}

	shard.Mu.Lock()
	defer shard.Mu.Unlock()

	list := shard.RPush[key]

	if start < 0 {
		start = len(list) + start
		if start < 0 {
			start = 0
		}
	}
	if end < 0 {
		end = len(list) + end
		if end < 0 {
			end = 0
		}
	}
	if start >= len(list) || start > end {
		Conn.Write([]byte("*0\r\n"))
		return
	}
	if end >= len(list) {
		end = len(list) - 1
	}
	sublist := list[start : end+1]
	Conn.Write([]byte(fmt.Sprintf("*%d\r\n", len(sublist))))
	for _, item := range sublist {
		utils.WriteBulkString(Conn, item)
	}
}
