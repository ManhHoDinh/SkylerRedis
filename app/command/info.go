package command

import (
	"fmt"
	"net"
	"skylerRedis/app/utils"
	"strings"
)

func handleINFO(conn net.Conn, args []string) {
	if len(args) != 2 {
		utils.WriteError(conn, "wrong number of arguments for 'INFO'")
		return
	}
	switch strings.ToUpper(args[1]) {
	case "REPLICATION":
		utils.WriteBulkString(conn, "role:master")
	default:
		utils.WriteError(conn, fmt.Sprintf("unknown INFO section '%s'", args[1]))
	}
}
