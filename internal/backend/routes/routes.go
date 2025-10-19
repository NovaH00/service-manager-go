package routes

import (
	"service-manager/docs"
	"service-manager/internal/manager"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func RegisterRoutes(router *gin.Engine, sm *manager.ServiceManager, logsDir string) {
	// programmatically set swagger info
	docs.SwaggerInfo.BasePath = "/"

	RegisterServiceManagerRoutes(router, sm)
	RegisterStreamRoutes(router, sm, logsDir)

	// Redirect /docs to /docs/
	router.GET("/docs", func(c *gin.Context) {
		c.Redirect(301, "/docs/")
	})
	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

