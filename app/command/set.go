package command

import (
	"SkylerRedis/app/memory"
	"SkylerRedis/app/server"
	"SkylerRedis/app/utils"
	"strconv"
	"strings"
	"time"
)

func handleSet(server server.Server, args []string) {
	if len(args) < 3 {
		utils.WriteError(server.Conn, "wrong number of arguments for 'SET'")
		return
	}
	key := args[1]
	val := args[2]
	var expiry time.Time

	if len(args) >= 5 && strings.ToUpper(args[3]) == "PX" {
		ms, err := strconv.Atoi(args[4])
		if err != nil {
			utils.WriteError(server.Conn, "PX value must be integer")
			return
		}
		expiry = time.Now().Add(time.Duration(ms) * time.Millisecond)
	}
	memory.Store[key] = memory.Entry{Value: val, ExpiryTime: expiry}
	utils.WriteSimpleString(server.Conn, "OK")
}
