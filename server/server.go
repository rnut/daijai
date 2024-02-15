package server

import (
	"daijai/config"
	"log"
)

func Init() {
	db := config.GetDB()
	r := SetupRouter(db)
	// config := config.GetConfig()
	// serverAddress := config.GetString("server.port")
	serverAddress := "0.0.0.0:8080"
	err := r.Run(serverAddress)
	if err != nil {
		log.Fatal(err)
	}
}
