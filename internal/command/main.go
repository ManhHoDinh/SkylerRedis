package command

import (
	"SkylerRedis/internal/memory"
	"SkylerRedis/internal/utils"
	"fmt"
	"net"
	"strings"
)

type ICommand interface {
	Handle(Conn net.Conn, args []string, isMaster bool, shard *memory.Shard)
}

func HandleCommand(Conn net.Conn, args []string, isMaster bool, shard *memory.Shard) { // Added shard parameter
	if len(args) == 1 && strings.ToUpper(args[0]) == "EXEC" {
		if !memory.IsMulti[Conn] {
			utils.WriteError(Conn, "EXEC without MULTI")
			return
		}
		memory.IsMulti[Conn] = false
		Conn.Write([]byte(fmt.Sprintf("*%d\r\n", len(memory.Queue[Conn]))))
		for _, cmd := range memory.Queue[Conn] {
			HandleCommand(Conn, cmd, isMaster, shard) // Pass shard here
		}
		memory.Queue[Conn] = nil
		return
	}
	if len(args) == 1 && strings.ToUpper(args[0]) == "DISCARD" {
		if !memory.IsMulti[Conn] {
			utils.WriteError(Conn, "DISCARD without MULTI")
			return
		}
		memory.IsMulti[Conn] = false
		memory.Queue[Conn] = nil
		utils.WriteSimpleString(Conn, "OK")
		return
	}

	if memory.IsMulti[Conn] {
		memory.Queue[Conn] = append(memory.Queue[Conn], args)
		utils.WriteSimpleString(Conn, "QUEUED")
		return
	} else {
		fmt.Println("Handling command:", args)
		fmt.Println("First command:", args[0])
		var cmd ICommand
		switch strings.ToUpper(args[0]) {
		case "PING":
			cmd = Ping{}
		case "ECHO":
			cmd = Echo{}
		case "SET":
			cmd = Set{}
		case "GET":
			cmd = Get{}
		case "LPUSH":
			cmd = LPush{}
		case "RPUSH":
			cmd = RPush{}
		case "LRANGE":
			cmd = LRange{}
		case "LLEN":
			cmd = LLen{}
		case "LPOP":
			cmd = LPop{}
		case "BLPOP":
			cmd = BLPop{}
		case "INCR":
			cmd = INCR{}
		case "MULTI":
			cmd = MULTI{}
		case "INFO":
			cmd = INFO{}
		case "REPLCONF":
			cmd = REPLCONF{}
		case "PSYNC":
			cmd = PSYNC{}
		case "TYPE":
			cmd = Type{}
		case "XADD":
			cmd = XAdd{}
		case "SADD":
			cmd = Sadd{}
		case "SCARD":
			cmd = Scard{}
		case "SISMEMBER":
			cmd = Sismember{}
		case "SMEMBERS":
			cmd = Smembers{}
		case "SREM":
			cmd = Srem{}
		case "BFADD":
			cmd = Bfadd{}
		case "BFEXISTS":
			cmd = Bfexists{}
		case "CMSINCRBY":
			cmd = Cmsincrby{}
		case "CMSQUERY":
			cmd = Cmsquery{}
		// case "XREAD":
		// 	handleXRead(server, args)
		default:
			utils.WriteError(Conn, fmt.Sprintf("unknown command '%s'", args[0]))
			return // Added return here
		}
		cmd.Handle(Conn, args, isMaster, shard) // Pass shard here
	}
}
