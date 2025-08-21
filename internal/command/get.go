package command

import (
	"SkylerRedis/internal/memory"
	"SkylerRedis/internal/utils"
	"fmt"
	"net"
	"time"
)

type Get struct{}

func (Get) Handle(Conn net.Conn, args []string, isMaster bool) {
	if len(args) != 2 {
		utils.WriteError(Conn, "wrong number of arguments for 'GET'")
		return
	}
	key := args[1]
	entry, ok := memory.Store[key]
	fmt.Println("GET", key)
	fmt.Println("Entry:", entry)
	if !ok || (entry.ExpiryTime != (time.Time{}) && time.Now().After(entry.ExpiryTime)) {
		delete(memory.Store, key)
		utils.WriteNull(Conn)
		return
	}
	utils.WriteBulkString(Conn, entry.Value)
}
