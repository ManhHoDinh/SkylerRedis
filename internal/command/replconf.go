package command

import (
	"SkylerRedis/internal/entity"
	"SkylerRedis/internal/memory"
	"SkylerRedis/internal/utils"
	"fmt"
	"net"
	"strconv"
	"strings"
)

type REPLCONF struct{}

func (REPLCONF) Handle(Conn net.Conn, args []string, isMaster bool, shard *memory.Shard) {
	fmt.Println("Handling REPLCONF command with args:", args)
	if strings.ToUpper(args[1]) == "LISTENING-PORT" {
		remoteHost, _, _ := net.SplitHostPort(Conn.RemoteAddr().String())
		fmt.Println("Remote host:", remoteHost)
		slaveAddr := utils.FormatAddr(remoteHost, args[2])
		fmt.Println("Adding slave with address:", slaveAddr)
		slave := entity.BaseServer{
			Conn: Conn,
			Addr: slaveAddr,
		}
		memory.Slaves = append(memory.Slaves, slave)
		fmt.Println("slaves:", memory.Slaves)
	}
	if strings.ToUpper(args[1]) == "GETACK" {
		fmt.Println("Handling REPLCONF GETACK command")
		ackItems := []string{
			"REPLCONF",
			"ACK",
			strconv.Itoa(memory.OffSet),
		}
		utils.WriteArray(Conn, ackItems)
		return
	}

	utils.WriteSimpleString(Conn, "OK")
}
