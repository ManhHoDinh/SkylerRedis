package command

import (
	"SkylerRedis/internal/memory"
	"SkylerRedis/internal/utils"
	"net"
	"time"
)

type Get struct{}

func (Get) Handle(Conn net.Conn, args []string, isMaster bool, masterReplID string, masterReplOffset int, connectedSlaves int, shard *memory.Shard) {
	if len(args) != 2 {
		utils.WriteError(Conn, "wrong number of arguments for 'GET'")
		return
	}
	key := args[1]

	shard.Mu.Lock()
	defer shard.Mu.Unlock()

	entryPtr, ok := shard.Store[key]
	// fmt.Println("GET", key) // Removed debug print
	// fmt.Println("Entry:", entryPtr) // Removed debug print
	if !ok || (entryPtr.ExpiryTime != (time.Time{}) && time.Now().After(entryPtr.ExpiryTime)) {
		delete(shard.Store, key)
		utils.WriteNull(Conn)
		return
	}

	// Update LRU and increment global LRU clock
	entryPtr.LRU = shard.LruClock
	shard.LruClock++
	// Since entryPtr is a pointer, modifying its fields directly changes the stored Entry.
	// Reassigning to map ensures any potential map internals are updated if Go ever requires it.
	shard.Store[key] = entryPtr 
	
	utils.WriteBulkString(Conn, entryPtr.Value)
}
