package command

import (
	"SkylerRedis/app/memory"
	"SkylerRedis/app/server"
	"SkylerRedis/app/utils"
)

func handleXAdd(server server.Server, args []string) {
	// Implementation for XADD command
	if len(args) < 4 {
		utils.WriteError(server.Conn, "ERR wrong number of arguments for 'XADD' command")
		return
	}
	var values = make(map[string]memory.Entry)
	for i := 3; i < len(args); i += 2 {
		values[args[i]] = memory.Entry{Value: args[i+1]}
	}
	memory.Stream[args[1]] = append(memory.Stream[args[1]], memory.StreamEntry{
		ID:    args[2],
		Values: values,
	})
	utils.WriteSimpleString(server.Conn, "OK")
}