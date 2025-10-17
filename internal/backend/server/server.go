package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"service-manager/internal/backend/routes"
	"service-manager/internal/manager"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Server struct {
	Router         *gin.Engine
	ServiceManager *manager.ServiceManager
	Host           string
	Port           string
}

func NewServer(logsDir, servicesDataPath, host, port string) (*Server, error) {
	// Server startup logics here

	serviceManager := manager.NewServiceManager(logsDir, servicesDataPath)

	err := serviceManager.LoadServices()
	if err != nil {
		log.Printf("could not load services from file: %v", err)
	}

	router := gin.Default()
	router.Use(cors.Default()) // Allow all origin
	router.HandleMethodNotAllowed = true

	routes.RegisterRoutes(router, serviceManager, logsDir)

	return &Server{
		Router:         router,
		ServiceManager: serviceManager,
		Host:           host,
		Port:           port,
	}, nil
}

func (s *Server) Run() {
	address := fmt.Sprintf("%s:%s", s.Host, s.Port)

	srv := &http.Server{
		Addr:    address,
		Handler: s.Router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}

	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Received shutdown signal")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	s.stop()

	log.Println("server exited")
}

func (s *Server) stop() {
	s.ServiceManager.StopAllServices()
}
