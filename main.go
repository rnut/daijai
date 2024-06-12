package main

import (
	"daijai/config"
	"daijai/server"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file", err)
	}
	mode := os.Getenv("GIN_MODE")
	if mode == gin.DebugMode {
		mode = gin.DebugMode
	} else {
		mode = gin.ReleaseMode
	}
	gin.SetMode(mode)
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
