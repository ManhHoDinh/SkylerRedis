package command

import (
	"SkylerRedis/app/memory"
	"SkylerRedis/app/server"
	"SkylerRedis/app/utils"
	"strconv"
	"time"
)

func handleBLPop(server server.Server, args []string) {
	if len(args) != 3 {
		utils.WriteError(server.Conn, "wrong number of arguments for 'BLPOP'")
		return
	}

	memory.Mu.Lock()
	key := args[1]
	if list, ok := memory.RPush[key]; ok && len(list) > 0 {
		value := list[0]
		memory.RPush[key] = list[1:]
		memory.Mu.Unlock()

		server.Conn.Write([]byte("*2\r\n"))
		utils.WriteBulkString(server.Conn, key)
		utils.WriteBulkString(server.Conn, value)
		return
	}

	timeoutStr := args[2]
	timeout, err := strconv.ParseFloat(timeoutStr, 64)
	if err != nil {
		utils.WriteError(server.Conn, "timeout must be a number")
		return
	}

	ch := make(chan string, 1)
	blocking := memory.BlockingRequest{
		Key:     key,
		Ch:      ch,
		Timeout: time.Duration(timeout * float64(time.Second)),
	}
	memory.Blockings[key] = append(memory.Blockings[key], blocking)
	memory.Mu.Unlock()

	if timeout == 0 {
		_, ok := <-ch
		if !ok {
			utils.WriteBulkString(server.Conn, "")
			return
		}
		list := memory.RPush[key]
		if len(list) > 0 {
			value := list[0]
			memory.RPush[key] = list[1:]
			server.Conn.Write([]byte("*2\r\n"))
			utils.WriteBulkString(server.Conn, key)
			utils.WriteBulkString(server.Conn, value)
			return
		}
	} else {
		select {
		case <-time.After(blocking.Timeout):
			memory.Mu.Lock()
			list := memory.Blockings[key]
			newList := []memory.BlockingRequest{}
			for _, r := range list {
				if r.Ch != ch {
					newList = append(newList, r)
				}
			}
			memory.Blockings[key] = newList
			memory.Mu.Unlock()
			utils.WriteNull(server.Conn)
			return
		case key := <-ch:
			list := memory.RPush[key]
			if len(list) > 0 {
				value := list[0]
				memory.RPush[key] = list[1:]
				server.Conn.Write([]byte("*2\r\n"))
				utils.WriteBulkString(server.Conn, key)
				utils.WriteBulkString(server.Conn, value)
				return
			}
		}
	}
}
