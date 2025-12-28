package command

import (
	"SkylerRedis/internal/memory"
	"SkylerRedis/internal/utils"
	"fmt"
	"net"
	"strconv"
	"strings"
)

type REPLCONF struct{}

func (REPLCONF) Handle(Conn net.Conn, args []string, isMaster bool, masterReplID string, masterReplOffset int, connectedSlaves int, shard *memory.Shard) {
	fmt.Println("Handling REPLCONF command with args:", args)

	if !isMaster {
		utils.WriteError(Conn, "ERR REPLCONF command only accepted by master server")
		return
	}

	if strings.ToUpper(args[1]) == "GETACK" {
		fmt.Println("Handling REPLCONF GETACK command")
		ackItems := []string{
			"REPLCONF",
			"ACK",
			strconv.Itoa(masterReplOffset), // Use master's offset
		}
		utils.WriteArray(Conn, ackItems)
		return
	}

	utils.WriteSimpleString(Conn, "OK")
}
