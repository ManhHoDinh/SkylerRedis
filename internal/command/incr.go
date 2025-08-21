package command

import (
	"SkylerRedis/internal/memory"
	"SkylerRedis/internal/utils"
	"net"
	"strconv"
	"time"
)

type INCR struct{}

func (INCR) Handle(Conn net.Conn, args []string, isMaster bool) {
	if len(args) != 2 {
		utils.WriteError(Conn, "wrong number of arguments for 'INCR'")
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
		utils.WriteError(Conn, "value is not an integer or out of range")
		return
	}
	val++
	memory.Store[key] = memory.Entry{Value: strconv.Itoa(val), ExpiryTime: entry.ExpiryTime}
	utils.WriteInteger(Conn, val)
}
