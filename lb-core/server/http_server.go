package server

import (
	"fmt"

	"github.com/CakeForKit/rsoi-lab2/lb-core/auth"
	"github.com/CakeForKit/rsoi-lab2/lb-core/config"
	"github.com/CakeForKit/rsoi-lab2/lb-core/controller"
	"github.com/gin-gonic/gin"
)

type Runnable interface {
	Run() error
}

type httpServer struct {
	router      *gin.Engine
	startupFunc func(*gin.Engine)
}

func NewHttpServer(startupFunc func(*gin.Engine)) Runnable {
	return &httpServer{
		startupFunc: startupFunc,
	}
}

func (server *httpServer) Run() error {
	server.router = gin.New()
	server.router.Use(gin.Recovery())
	controller.NewHealthCheckController().RegisterHttpController(server.router)
	server.router.Use(auth.Middleware())

	server.startupFunc(server.router)
	return server.router.Run(fmt.Sprintf(":%d", config.CoreConfig.Port))
}
