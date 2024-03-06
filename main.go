package main

import (
	"daijai/config"
	"daijai/server"

	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.DebugMode)
	// gin.SetMode(gin.ReleaseMode)
	// environment := flag.String("e", "dev", "")
	// flag.Usage = func() {
	// 	fmt.Println("Usage: server -e {mode}")
	// 	os.Exit(1)
	// }
	// flag.Parse()
	// config.Init(*environment)
	config.ConnectDB()
	server.Init()
}
