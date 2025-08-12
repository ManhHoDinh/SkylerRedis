package command

import (
	"SkylerRedis/app/memory"
	"SkylerRedis/app/server"
	"SkylerRedis/app/utils"
	"fmt"
)

func handleREPLCONF(request server.Server, args []string) {
	fmt.Println("Handling REPLCONF command with args:", args)
	go func ()  {
		if(args[1] == "listening-network"){
			ser := server.Server{
					Addr: args[2],
				}
			slave := server.Slave{
				Server: &ser,
			}
			fmt.Println("Adding slave:", slave.Server.Addr)
			memory.Master.Slaves = append(memory.Master.Slaves, &slave)
			fmt.Println("slaves:", memory.Master.Slaves)
		}	
	}()
	utils.WriteSimpleString(request.Conn, "OK")
}