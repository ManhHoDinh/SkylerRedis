package handler

import (
	"SkylerRedis/app/command"
	"SkylerRedis/app/memory"
	"SkylerRedis/app/server"
	"SkylerRedis/app/utils"
	"bufio"
	"fmt"
	"net"
	"strings"
)

func argsToRESP(args []string) []byte {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("*%d\r\n", len(args)))
	for _, arg := range args {
		b.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(arg), arg))
	}
	return []byte(b.String())
}

func HandleConnection(requestServer server.Server) {
	reader := bufio.NewReader(requestServer.Conn)
	defer requestServer.Conn.Close()
	for {
		args, err := utils.ParseArgs(requestServer.Conn, reader)
		if err != nil {
			utils.WriteError(requestServer.Conn, err.Error())
			continue
		}
		if len(args) == 0 {
			utils.WriteError(requestServer.Conn, "empty command")
			return
		}

		command.HandleCommand(requestServer, args)

		if requestServer.IsMaster && isModifyCommand(args) {
			fmt.Println("Master detected, forwarding command to slaves")
			// Forward command to all slaves
			fmt.Println("Forwarding to slaves:", len(memory.Master.Slaves))

			resp := argsToRESP(args)
			for _, ser := range memory.Master.Slaves {
				conn, err := net.Dial("tcp", ser.Addr)
				if err != nil {
					fmt.Println("Error connecting to slave:", err)
					fmt.Println("Retrying connect to slave:", ser.Conn.RemoteAddr())
					_, reTryErr := ser.Conn.Write(resp)
					if reTryErr != nil {
						fmt.Println("Error forwarding to slave:", reTryErr)
					}
					fmt.Println("Command forwarded to slave:", ser.Conn.RemoteAddr())
					continue
				}
				fmt.Println("Forwarding to slave:", ser.Addr)
				_, err = conn.Write(resp)
				if err != nil {
					fmt.Println("Error forwarding to slave:", err)
				}
				fmt.Println("Command forwarded to slave:", ser.Addr)
				defer conn.Close()
			}
		}
	}
}

func isModifyCommand(args []string) bool {
	return len(args) > 0 && (args[0] == "SET" || args[0] == "DEL" || args[0] == "LPUSH" || args[0] == "RPUSH" || args[0] == "INCR")
}
