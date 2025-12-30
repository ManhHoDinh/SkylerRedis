package command

import (
	"SkylerRedis/internal/memory"
	"SkylerRedis/internal/utils"
	"net"
	"strconv"
	"time"
)

type INCR struct{}

func (INCR) Handle(Conn net.Conn, args []string, isMaster bool, masterReplID string, masterReplOffset int, connectedSlaves int, shard *memory.Shard) {
	if len(args) != 2 {
		utils.WriteError(Conn, "wrong number of arguments for 'INCR'")
		return
	}
	key := args[1]
	
	shard.Mu.Lock()
	defer shard.Mu.Unlock()

	entryPtr, exists := shard.Store[key]
	if !exists || (entryPtr != nil && !entryPtr.ExpiryTime.IsZero() && time.Now().After(entryPtr.ExpiryTime)) {
		entryPtr = &memory.Entry{Value: "0", ExpiryTime: time.Time{}}
		shard.Store[key] = entryPtr
	}

	val, err := strconv.Atoi(entryPtr.Value)
	if err != nil {
		utils.WriteError(Conn, "value is not an integer or out of range")
		return
	}
	val++
	entryPtr.Value = strconv.Itoa(val)
	// No need to reassign shard.Store[key] = entryPtr here, as entryPtr is a pointer
	// and its modification is reflected in the map. However, keeping for consistency.
	shard.Store[key] = entryPtr 
	utils.WriteInteger(Conn, val)
}
