package main

import (
	"daijai/config"
	"daijai/server"
	"flag"
	"fmt"
	"os"
)

func main() {
	environment := flag.String("e", "dev", "")
	flag.Usage = func() {
		fmt.Println("Usage: server -e {mode}")
		os.Exit(1)
	}
	flag.Parse()
	config.Init(*environment)
	config.ConnectDB()
	server.Init()
}
