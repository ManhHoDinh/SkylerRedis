package command

import (
	"SkylerRedis/internal/memory"
	"SkylerRedis/internal/utils"
	"net"
)

type XAdd struct{}

func (XAdd) Handle(Conn net.Conn, args []string, isMaster bool) {
	if len(args) < 4 {
		utils.WriteError(Conn, "ERR wrong number of arguments for 'XADD' command")
		return
	}
	if len(memory.StreamIDs) > 0 {
		if memory.StreamIDs[len(memory.StreamIDs)-1] >= args[2] {
			utils.WriteError(Conn, "ERR The ID specified in XADD is equal or smaller than the target stream top item")
			return
		}
	}
	var values = make(map[string]memory.Entry)
	for i := 3; i < len(args); i += 2 {
		values[args[i]] = memory.Entry{Value: args[i+1]}
	}
	memory.Stream[args[2]] = memory.StreamEntry{
		ID:     args[2],
		Values: values,
	}
	memory.StreamIDs = append(memory.StreamIDs, args[2])
	utils.WriteSimpleString(Conn, args[2])
}
