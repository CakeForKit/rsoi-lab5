package main

import (
	"fmt"

	"github.com/CakeForKit/rsoi-lab2/lb-core/config"
	"github.com/CakeForKit/rsoi-lab2/lb-core/server"
	"github.com/CakeForKit/rsoi-lab2/lb-core/utils"
	"github.com/CakeForKit/rsoi-lab2/ms-ticket/internal/controller"
	"github.com/gin-gonic/gin"
)

func main() {
	utils.InitLogger()

	utils.CheckedError(config.Load(&config.CoreConfig))

	fmt.Printf("config: %d\n\n", config.CoreConfig.Port)

	utils.CheckedError(server.NewHttpServer(func(router *gin.Engine) {
		utils.RegisterController(router, controller.GetTicketController)
	}).Run())
}
