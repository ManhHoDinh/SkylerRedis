package command

import (
	"SkylerRedis/app/memory"
	"SkylerRedis/app/server"
	"SkylerRedis/app/utils"
	"fmt"
	"strings"
)

func HandleCommand(server server.Server, args []string) {
	if len(args) == 1 && strings.ToUpper(args[0]) == "EXEC" {
		if !memory.IsMulti[server.Conn] {
			utils.WriteError(server.Conn, "EXEC without MULTI")
			return
		}
		memory.IsMulti[server.Conn] = false
		server.Conn.Write([]byte(fmt.Sprintf("*%d\r\n", len(memory.Queue[server.Conn]))))
		for _, cmd := range memory.Queue[server.Conn] {
			HandleCommand(server, cmd)
		}
		memory.Queue[server.Conn] = nil
		return
	}
	if len(args) == 1 && strings.ToUpper(args[0]) == "DISCARD" {
		if !memory.IsMulti[server.Conn] {
			utils.WriteError(server.Conn, "DISCARD without MULTI")
			return
		}
		memory.IsMulti[server.Conn] = false
		memory.Queue[server.Conn] = nil
		utils.WriteSimpleString(server.Conn, "OK")
		return
	}

	if memory.IsMulti[server.Conn] {
		memory.Queue[server.Conn] = append(memory.Queue[server.Conn], args)
		utils.WriteSimpleString(server.Conn, "QUEUED")
		return
	} else {
		fmt.Println("Handling command:", args)
		fmt.Println("First command:", args[0])

		switch strings.ToUpper(args[0]) {
		case "PING":
			handlePing(server)
		case "ECHO":
			handleEcho(server, args)
		case "SET":
			handleSet(server, args)
		case "GET":
			handleGet(server, args)
		case "LPUSH":
			handleLPush(server, args)
		case "RPUSH":
			handleRPush(server, args)
		case "LRANGE":
			handleLRange(server, args)
		case "LLEN":
			handleLLen(server, args)
		case "LPOP":
			handleLPop(server, args)
		case "BLPOP":
			handleBLPop(server, args)
		case "INCR":
			handleINCR(server, args)
		case "MULTI":
			handleMULTI(server, args)
		case "INFO":
			handleINFO(server.Conn, args, server)
		case "REPLCONF":
			handleREPLCONF(server, args)
		case "PSYNC":
			handlePSYNC(server, args)
		case "TYPE":
			handleType(server, args)
		default:
			utils.WriteError(server.Conn, fmt.Sprintf("unknown command '%s'", args[0]))
		}
	}
}
