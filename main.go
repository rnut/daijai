package main

import (
	"daijai/config"
	"daijai/server"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file", err)
	}
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
