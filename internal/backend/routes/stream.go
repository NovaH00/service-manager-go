package routes

import (
	"service-manager/internal/backend/handlers"
	"service-manager/internal/manager"

	"github.com/gin-gonic/gin"
)

func RegisterStreamRoutes(router *gin.Engine, sm *manager.ServiceManager, logsDir string) {
	handler := handlers.NewStreamHandler(sm, logsDir)

	streamGroup := router.Group("/stream")
	{
		streamGroup.GET("/stdout/:serviceID", handler.StreamStdout)
		streamGroup.GET("/stderr/:serviceID", handler.StreamStderr)
	}
}
