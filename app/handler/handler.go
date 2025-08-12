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
// Chuyển args thành RESP bytes
func argsToRESP(args []string) []byte {
    var b strings.Builder
    b.WriteString(fmt.Sprintf("*%d\r\n", len(args)))
    for _, arg := range args {
        b.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(arg), arg))
    }
    return []byte(b.String())
}

func HandleConnection(requestServer server.Server) {
    defer requestServer.Conn.Close()
    reader := bufio.NewReader(requestServer.Conn)

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

        if requestServer.IsMaster {
            fmt.Println("Master detected, forwarding command to slaves")
            for _, ser := range memory.Master.Slaves {
                fmt.Println("Forwarding command to slave:", ser.Addr)
                conn, err := net.Dial("tcp", ser.Addr)
                if err != nil {
                    fmt.Println("Error connecting to slave:", err)
                    continue
                }
                defer conn.Close()

                resp := argsToRESP(args)
                conn.Write(resp)
            }
        }
    }
}

// Helpers
