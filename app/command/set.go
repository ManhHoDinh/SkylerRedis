package command

import (
	"net"
	"skylerRedis/app/memory"
	"skylerRedis/app/utils"
	"strconv"
	"strings"
	"time"
)

func handleSet(conn net.Conn, args []string) {
	if len(args) < 3 {
		utils.WriteError(conn, "wrong number of arguments for 'SET'")
		return
	}
	key := args[1]
	val := args[2]
	var expiry time.Time

	if len(args) >= 5 && strings.ToUpper(args[3]) == "PX" {
		ms, err := strconv.Atoi(args[4])
		if err != nil {
			utils.WriteError(conn, "PX value must be integer")
			return
		}
		expiry = time.Now().Add(time.Duration(ms) * time.Millisecond)
	}

	memory.Store[key] = memory.Entry{Value: val, ExpiryTime: expiry}
	utils.WriteSimpleString(conn, "OK")
}
