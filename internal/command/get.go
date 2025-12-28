package command

import (
	"SkylerRedis/internal/memory"
	"SkylerRedis/internal/utils"
	"fmt"
	"net"
	"time"
)

type Get struct{}

func (Get) Handle(Conn net.Conn, args []string, isMaster bool, shard *memory.Shard) {
	if len(args) != 2 {
		utils.WriteError(Conn, "wrong number of arguments for 'GET'")
		return
	}
	key := args[1]

	shard.Mu.Lock()
	defer shard.Mu.Unlock()

	entry, ok := shard.Store[key]
	fmt.Println("GET", key)
	fmt.Println("Entry:", entry)
	if !ok || (entry.ExpiryTime != (time.Time{}) && time.Now().After(entry.ExpiryTime)) {
		delete(shard.Store, key)
		utils.WriteNull(Conn)
		return
	}

	// Update LRU and increment global LRU clock
	entry.LRU = shard.LruClock
	shard.LruClock++
	shard.Store[key] = entry // Write back the updated entry
	
	utils.WriteBulkString(Conn, entry.Value)
}
