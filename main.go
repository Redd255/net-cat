package main

import (
	"fmt"
	serverNet "netcat/src"
	"os"
	"strconv"
)

func main() {
	port := ":8989"
	if len(os.Args) == 2 {
		if _, err := strconv.Atoi(os.Args[1]); err == nil {
			port = ":" + os.Args[1]
		} else {
			fmt.Println("[USAGE]: ./TCPChat $port")
			return
		}
	} else if len(os.Args) > 2 {
		fmt.Println("[USAGE]: ./TCPChat $port")
		return
	}

	server := serverNet.NewServer(port)
	server.Start()
}
