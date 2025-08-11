package handler

import (
	"SkylerRedis/app/command"
	"SkylerRedis/app/memory"
	"SkylerRedis/app/server"
	"SkylerRedis/app/utils"
	"bufio"
	"fmt"
	"net"
	"time"
)

func HandleConnection(conn net.Conn, server server.Server) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	for {
		args, err := utils.ParseArgs(conn, reader)
		if err != nil {
			utils.WriteError(conn, err.Error())
			continue
		}
		if len(args) == 0 {
			utils.WriteError(conn, "empty command")
			return
		}
		command.HandleCommand(conn, args, server)
		if server.IsMaster {
			fmt.Println("Master detected, forwarding command to slaves")

			for _, ser := range memory.Master.Slaves {
				fmt.Println("Forwarding command to slave:", ser.Addr)
				slaveConn, err := net.Dial("tcp", ser.Addr)
				if err != nil {
					fmt.Println("Failed to connect to ", ser.Addr, err)
					time.Sleep(5 * time.Second)
				}
				command.HandleCommand(slaveConn, args, *ser.Server)
			}
		}
	}
}

// Helpers
