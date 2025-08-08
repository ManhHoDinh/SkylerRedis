package command

import (
	"SkylerRedis/app/memory"
	"SkylerRedis/app/utils"
	"net"
	"strconv"
	"time"
)

func handleINCR(conn net.Conn, args []string) {
	if len(args) != 2 {
		utils.WriteError(conn, "wrong number of arguments for 'INCR'")
		return
	}
	key := args[1]
	entry, exists := memory.Store[key]
	if !exists || entry.ExpiryTime != (time.Time{}) && time.Now().After(entry.ExpiryTime) {
		memory.Store[key] = memory.Entry{Value: "0", ExpiryTime: time.Time{}}
		entry = memory.Store[key]
	}

	val, err := strconv.Atoi(entry.Value)
	if err != nil {
		utils.WriteError(conn, "value is not an integer or out of range")
		return
	}
	val++
	memory.Store[key] = memory.Entry{Value: strconv.Itoa(val), ExpiryTime: entry.ExpiryTime}
	utils.WriteInteger(conn, val)
}
