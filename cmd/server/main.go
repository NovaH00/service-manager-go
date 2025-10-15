// @title           Service Manager API
// @version         1.0
// @description     An API for managing and monitoring background services.
// @host            localhost:8080
// @BasePath        /
package main

import (
	"log"
	"service-manager/internal/backend/server"
	"service-manager/internal/backend/utils"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment variables")
	}
	HOST := utils.GetEnv("HOST", "0.0.0.0")
	PORT := utils.GetEnv("PORT", "8080")
	LOGS_DIR := utils.GetEnv("LOGS_DIR", "data/logs")
	SERVICES_DATA := utils.GetEnv("SERVICES_DATA", "data/services_data.json")

	srv, err := server.NewServer(LOGS_DIR, SERVICES_DATA, HOST, PORT)
	if err != nil {
		log.Println("create server: ", err)
	}

	srv.Run()
}
