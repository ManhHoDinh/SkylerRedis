package command

import (
	"SkylerRedis/internal/memory"
	"SkylerRedis/internal/utils"
	"net"
	"strconv"
	"strings"
	"time"
)

type Set struct{}

func (Set) Handle(Conn net.Conn, args []string, isMaster bool) {
	if len(args) < 3 {
		utils.WriteError(Conn, "wrong number of arguments for 'SET'")
		return
	}
	key := args[1]
	val := args[2]
	var expiry time.Time

	if len(args) >= 5 && strings.ToUpper(args[3]) == "PX" {
		ms, err := strconv.Atoi(args[4])
		if err != nil {
			utils.WriteError(Conn, "PX value must be integer")
			return
		}
		expiry = time.Now().Add(time.Duration(ms) * time.Millisecond)
	}
	memory.Store[key] = memory.Entry{Value: val, ExpiryTime: expiry}
	utils.WriteSimpleString(Conn, "OK")
}
