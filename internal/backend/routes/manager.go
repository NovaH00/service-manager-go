package routes

import (
	"service-manager/internal/backend/handlers"
	"service-manager/internal/manager"

	"github.com/gin-gonic/gin"
)

func RegisterServiceManagerRoutes(router *gin.Engine, sm *manager.ServiceManager) {
	handler := handlers.NewServiceManagerHandler(sm)

	serviceManagerGroup := router.Group("/manager")
	{
		serviceManagerGroup.POST("/register", handler.RegisterService)
		serviceManagerGroup.GET("/services", handler.GetServices)
		serviceManagerGroup.POST("/start", handler.StartService)
		serviceManagerGroup.POST("/stop", handler.StopService)
		serviceManagerGroup.DELETE("/remove", handler.RemoveService)
		serviceManagerGroup.POST("/metrics", handler.GetServiceMetrics)
		serviceManagerGroup.POST("/network", handler.GetNetworkInfo)
	}

}
