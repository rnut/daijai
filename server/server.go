package server

import (
	"daijai/config"
	"log"
	"os"
)

func Init() {
	db := config.GetDB()
	r := SetupRouter(db)
	// config := config.GetConfig()
	// serverAddress := config.GetString("server.port")
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "keys/daijai-d4ab4aa6981d.json")
	serverAddress := "0.0.0.0:8080"
	err := r.Run(serverAddress)
	if err != nil {
		log.Fatal(err)
	}
}
