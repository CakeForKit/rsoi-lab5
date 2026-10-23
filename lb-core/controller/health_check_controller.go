package controller

import (
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
)

type healthCheckController struct{}

func NewHealthCheckController() HttpController {
	return &healthCheckController{}
}

func (controller *healthCheckController) RegisterHttpController(router *gin.Engine) {
	healthRouter := router.Group("/manage/health")
	healthRouter.GET("", controller.getHealth)
}

func (controller *healthCheckController) getHealth(ctx *gin.Context) {
	log.WithContext(ctx).Debug("GetHealth: Start")
	ctx.JSON(http.StatusOK, gin.H{})
	log.WithContext(ctx).Debug("GetHealth: Success")
}
