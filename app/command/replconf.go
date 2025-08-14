package command

import (
	"SkylerRedis/app/memory"
	"SkylerRedis/app/server"
	"SkylerRedis/app/utils"
	"fmt"
	"net"
	"strconv"
	"strings"
)

func handleREPLCONF(request server.Server, args []string) {
	fmt.Println("Handling REPLCONF command with args:", args)
	if strings.ToUpper(args[1]) == "LISTENING-PORT" {
		remoteHost, _, _ := net.SplitHostPort(request.Conn.RemoteAddr().String())
		fmt.Println("Remote host:", remoteHost)
		slaveAddr := utils.FormatAddr(remoteHost, args[2])
		fmt.Println("Adding slave with address:", slaveAddr)
		ser := server.Server{
			Conn: request.Conn,
			Addr: slaveAddr,
		}
		slave := server.Slave{
			Server: &ser,
			Master: memory.Master,
		}
		memory.Master.Slaves = append(memory.Master.Slaves, &slave)
		fmt.Println("slaves:", memory.Master.Slaves)
	}
	if strings.ToUpper(args[1]) == "GETACK" {
		fmt.Println("Handling REPLCONF GETACK command")
		ackItems := []string{
			"REPLCONF",
			"ACK",
			strconv.Itoa(memory.OffSet),
		}
		memory.OffSet += 37
		utils.WriteArray(request.Conn, ackItems)
		return
	}
	
	utils.WriteSimpleString(request.Conn, "OK")
}
