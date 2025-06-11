package server

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
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

	s.registerRoutes(s.router.Group("/api/v1"))
	s.registerRoutes(s.router.Group("/v1"))
}

func (s *Server) registerRoutes(r *gin.RouterGroup) {
	// farms routes
	publicFarmRoutes := r.Group("farms")
	publicFarmRoutes.GET("", s.listFarmsHandler)
	publicFarmRoutes.GET("/:farm_id", s.getFarmHandler)
	// protected by farmer key
	protectedFarmRoutes := r.Group("farms", s.AuthMiddleware())
	protectedFarmRoutes.POST("", s.createFarmHandler)
	// added to stop redirecting when creating a farm with extra /
	protectedFarmRoutes.POST("/", s.createFarmHandler)
	protectedFarmRoutes.PATCH("/:farm_id", s.updateFarmHandler)
	protectedFarmRoutes.POST("/:farm_id/approve", s.approveNodesHandler)

	// nodes routes
	publicNodeRoutes := r.Group("nodes")
	publicNodeRoutes.GET("", s.listNodesHandler)
	publicNodeRoutes.GET("/:node_id", s.getNodeHandler)
	// protected by node key
	protectedNodeRoutes := r.Group("nodes", s.AuthMiddleware())
	protectedNodeRoutes.POST("", s.registerNodeHandler)
	protectedNodeRoutes.PATCH("/:node_id", s.updateNodeHandler)
	protectedNodeRoutes.POST("/:node_id/uptime", s.uptimeReportHandler)

	// Account routes
	publicAccountRoutes := r.Group("accounts")
	// added to stop redirecting when creating an account with extra /
	publicAccountRoutes.POST("", s.createAccountHandler)
	publicAccountRoutes.POST("/", s.createAccountHandler)
	publicAccountRoutes.GET("", s.getAccountHandler)
	publicAccountRoutes.GET("/", s.getAccountHandler)
	// protected by farmer key
	protectedAccountRoutes := r.Group("accounts", s.AuthMiddleware())
	protectedAccountRoutes.PATCH("/:twin_id", s.updateAccountHandler)

	// zOS Version endpoints
	publicZosRoutes := r.Group("/zos")
	publicZosRoutes.GET("/version", s.getZOSVersionHandler)
	// protected by admin key
	protectedZosRoutes := r.Group("/zos", s.AuthMiddleware())
	protectedZosRoutes.PUT("/version", s.setZOSVersionHandler)
}
