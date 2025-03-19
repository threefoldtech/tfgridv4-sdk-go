package server

import (
	"time"

	"github.com/gin-contrib/cors"
	_ "github.com/threefoldtech/tfgrid4-sdk-go/node-registrar/docs"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func (s *Server) SetupRoutes() {
	s.router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"POST", "OPTIONS", "GET", "PUT", "DELETE"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"*"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	s.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	v1 := s.router.Group("/api/v1")

	// farms routes
	publicFarmRoutes := v1.Group("farms")
	publicFarmRoutes.GET("/", s.listFarmsHandler)
	publicFarmRoutes.GET("/:farm_id", s.getFarmHandler)
	// protected by farmer key
	protectedFarmRoutes := v1.Group("farms", s.AuthMiddleware())
	protectedFarmRoutes.POST("/", s.createFarmHandler)
	protectedFarmRoutes.PATCH("/:farm_id", s.updateFarmHandler)

	// nodes routes
	publicNodeRoutes := v1.Group("nodes")
	publicNodeRoutes.GET("/", s.listNodesHandler)
	publicNodeRoutes.GET("/:node_id", s.getNodeHandler)
	// protected by node key
	protectedNodeRoutes := v1.Group("nodes", s.AuthMiddleware())
	protectedNodeRoutes.POST("/", s.registerNodeHandler)
	protectedNodeRoutes.PATCH("/:node_id", s.updateNodeHandler)
	protectedNodeRoutes.POST("/:node_id/uptime", s.uptimeReportHandler)

	// Account routes
	publicAccountRoutes := v1.Group("accounts")
	publicAccountRoutes.POST("/", s.createAccountHandler)
	publicAccountRoutes.GET("/", s.getAccountHandler)
	// protected by farmer key
	protectedAccountRoutes := v1.Group("accounts", s.AuthMiddleware())
	protectedAccountRoutes.PATCH("/:twin_id", s.updateAccountHandler)

	// zOS Version endpoints
	publicZosRoutes := v1.Group("/zos")
	publicZosRoutes.GET("/version", s.getZOSVersionHandler)
	// protected by admin key
	protectedZosRoutes := v1.Group("/zos", s.AuthMiddleware())
	protectedZosRoutes.PUT("/version", s.setZOSVersionHandler)
}
