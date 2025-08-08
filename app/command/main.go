package command

import (
	"fmt"
	"net"
	"skylerRedis/app/memory"
	"skylerRedis/app/utils"
	"strings"
)

func HandleCommand(conn net.Conn, args []string) {
	if len(args) == 1 && strings.ToUpper(args[0]) == "EXEC" {
		if !memory.IsMulti[conn] {
			utils.WriteError(conn, "EXEC without MULTI")
			return
		}
		memory.IsMulti[conn] = false
		conn.Write([]byte(fmt.Sprintf("*%d\r\n", len(memory.Queue[conn]))))
		for _, cmd := range memory.Queue[conn] {
			HandleCommand(conn, cmd)
		}
		memory.Queue[conn] = nil
		return
	}
	if len(args) == 1 && strings.ToUpper(args[0]) == "DISCARD" {
		if !memory.IsMulti[conn] {
			utils.WriteError(conn, "DISCARD without MULTI")
			return
		}
		memory.IsMulti[conn] = false
		memory.Queue[conn] = nil
		utils.WriteSimpleString(conn, "OK")
		return
	}

	if memory.IsMulti[conn] {
		memory.Queue[conn] = append(memory.Queue[conn], args)
		utils.WriteSimpleString(conn, "memory.QueueD")
		return
	} else {
		switch strings.ToUpper(args[0]) {
		case "PING":
			handlePing(conn)
		case "ECHO":
			handleEcho(conn, args)
		case "SET":
			handleSet(conn, args)
		case "GET":
			handleGet(conn, args)
		case "LPUSH":
			handleLPush(conn, args)
		case "RPUSH":
			handleRPush(conn, args)
		case "LRANGE":
			handleLRange(conn, args)
		case "LLEN":
			handleLLen(conn, args)
		case "LPOP":
			handleLPop(conn, args)
		case "BLPOP":
			handleBLPop(conn, args)
		case "INCR":
			handleINCR(conn, args)
		case "MULTI":
			handleMULTI(conn, args)
		case "INFO":
			handleINFO(conn, args)
		default:
			utils.WriteError(conn, fmt.Sprintf("unknown command '%s'", args[0]))
		}
	}
}
