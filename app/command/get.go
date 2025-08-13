package command

import (
	"SkylerRedis/app/memory"
	"SkylerRedis/app/server"
	"SkylerRedis/app/utils"
	"fmt"
	"time"
)

func handleGet(server server.Server, args []string) {
	if len(args) != 2 {
		utils.WriteError(server.Conn, "wrong number of arguments for 'GET'")
		return
	}
	key := args[1]
	entry, ok := memory.Store[key]
	time.Sleep(5 * time.Millisecond)
	fmt.Println("GET", key)
	fmt.Println("Entry:", entry)
	if !ok || (entry.ExpiryTime != (time.Time{}) && time.Now().After(entry.ExpiryTime)) {
		delete(memory.Store, key)
		utils.WriteNull(server.Conn)
		return
	}
	utils.WriteBulkString(server.Conn, entry.Value)
}
