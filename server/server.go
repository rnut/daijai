package server

import (
	"daijai/config"
	"log"
)

func Init() {
	db := config.GetDB()
	config := config.GetConfig()
	r := SetupRouter(db)
	err := r.Run(config.GetString("server.port"))
	if err != nil {
		log.Fatal(err)
	}
}
