package command

import (
	"SkylerRedis/app/memory"
	"SkylerRedis/app/utils"
	"net"
	"time"
)

func handleGet(conn net.Conn, args []string) {
	if len(args) != 2 {
		utils.WriteError(conn, "wrong number of arguments for 'GET'")
		return
	}
	key := args[1]
	entry, ok := memory.Store[key]
	if !ok || (entry.ExpiryTime != (time.Time{}) && time.Now().After(entry.ExpiryTime)) {
		delete(memory.Store, key)
		utils.WriteNull(conn)
		return
	}
	utils.WriteBulkString(conn, entry.Value)
}
