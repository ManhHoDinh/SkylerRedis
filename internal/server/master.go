package server

import (
	"SkylerRedis/internal/command"
	"SkylerRedis/internal/entity"
	"SkylerRedis/internal/memory"
	"SkylerRedis/internal/utils"
	"bufio"
	"fmt"
	"net"
)

type Master struct {
	*entity.BaseServer
}

func (Master) HandShake(replicaof *string, port *string) {
	panic("Master doesn't make a handshark")
}
func (r Master) HandleConnection(Conn net.Conn) {
	reader := bufio.NewReader(Conn)
	for {
		args, err := utils.ParseArgs(Conn, reader)
		if err != nil {
			return
		}
		go command.HandleCommand(Conn, args, true)
		if utils.IsModifyCommand(args) {
			// Forward command to all slaves
			go forwardCommand(args)
		}
	}
}
func forwardCommand(args []string) {
	resp := utils.ArgsToRESP(args)
	fmt.Println(memory.Slaves)
	for _, ser := range memory.Slaves {
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
		_, err = conn.Write(resp)
		if err != nil {
			fmt.Println("Error forwarding to slave:", err)
		}
		fmt.Println("Command forwarded to slave:", ser.Addr)
	}
}
