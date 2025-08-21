package command

import (
	"SkylerRedis/internal/memory"
	"SkylerRedis/internal/utils"
	"net"
	"strconv"
	"time"
)

type BLPop struct{}

func (BLPop) Handle(Conn net.Conn, args []string, isMaster bool) {
	if len(args) != 3 {
		utils.WriteError(Conn, "wrong number of arguments for 'BLPOP'")
		return
	}

	memory.Mu.Lock()
	key := args[1]
	if list, ok := memory.RPush[key]; ok && len(list) > 0 {
		value := list[0]
		memory.RPush[key] = list[1:]
		memory.Mu.Unlock()
		utils.WriteArray(Conn, []string{key, value})
		return
	}

	timeoutStr := args[2]
	timeout, err := strconv.ParseFloat(timeoutStr, 64)
	if err != nil {
		utils.WriteError(Conn, "timeout must be a number")
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
			utils.WriteBulkString(Conn, "")
			return
		}
		list := memory.RPush[key]
		if len(list) > 0 {
			value := list[0]
			memory.RPush[key] = list[1:]
			Conn.Write([]byte("*2\r\n"))
			utils.WriteBulkString(Conn, key)
			utils.WriteBulkString(Conn, value)
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
			utils.WriteNull(Conn)
			return
		case key := <-ch:
			list := memory.RPush[key]
			if len(list) > 0 {
				value := list[0]
				memory.RPush[key] = list[1:]
				Conn.Write([]byte("*2\r\n"))
				utils.WriteBulkString(Conn, key)
				utils.WriteBulkString(Conn, value)
				return
			}
		}
	}
}
