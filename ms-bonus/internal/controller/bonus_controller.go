package controller

import (
	"github.com/CakeForKit/rsoi-lab2/lb-core/auth"
	coreController "github.com/CakeForKit/rsoi-lab2/lb-core/controller"
	"github.com/CakeForKit/rsoi-lab2/lb-core/model"
	"github.com/CakeForKit/rsoi-lab2/lb-core/utils"
	"github.com/CakeForKit/rsoi-lab2/ms-bonus/internal/service"
	"github.com/gin-gonic/gin"
	"net/http"
	"sync"
)

var bonusCntrlr coreController.HttpController
var bonusCntrlrMutex sync.Mutex

type bonusController struct{ service service.BonusService }

func GetBonusController() (coreController.HttpController, error) {
	bonusCntrlrMutex.Lock()
	defer bonusCntrlrMutex.Unlock()
	if bonusCntrlr != nil {
		return bonusCntrlr, nil
	}
	srv, err := service.GetBonusService()
	if err != nil {
		return nil, err
	}
	bonusCntrlr = &bonusController{service: srv}
	return bonusCntrlr, nil
}
func (c *bonusController) RegisterHttpController(router *gin.Engine) {
	router.GET("/internal/bonus", c.get)
	router.POST("/internal/bonus/operations", c.apply)
}
func (c *bonusController) get(ctx *gin.Context) {
	username := auth.Username(ctx)
	if username == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
		return
	}
	result, err := c.service.Get(ctx, username)
	if err != nil {
		utils.AbortContextWithError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, result)
}
func (c *bonusController) apply(ctx *gin.Context) {
	var operation model.PrivilegeOperationRequest
	if err := ctx.ShouldBindJSON(&operation); err != nil {
		utils.AbortContextWithError(ctx, err)
		return
	}
	result, err := c.service.Apply(ctx, auth.Username(ctx), operation)
	if err != nil {
		utils.AbortContextWithError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, result)
}
