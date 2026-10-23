package controller

import (
	"net/http"
	"strconv"
	"sync"

	coreController "github.com/CakeForKit/rsoi-lab2/lb-core/controller"
	"github.com/CakeForKit/rsoi-lab2/lb-core/utils"
	"github.com/CakeForKit/rsoi-lab2/ms-flight/internal/service"
	"github.com/gin-gonic/gin"
)

var flightCntrlr coreController.HttpController
var flightCntrlrMutex sync.Mutex

type flightController struct{ service service.FlightService }

func GetFlightController() (coreController.HttpController, error) {
	flightCntrlrMutex.Lock()
	defer flightCntrlrMutex.Unlock()
	if flightCntrlr != nil {
		return flightCntrlr, nil
	}
	srv, err := service.GetFlightService()
	if err != nil {
		return nil, err
	}
	flightCntrlr = &flightController{service: srv}
	return flightCntrlr, nil
}
func (c *flightController) RegisterHttpController(router *gin.Engine) {
	router.GET("/internal/flights", c.list)
	router.GET("/internal/flights/:flightNumber", c.get)
}
func (c *flightController) list(ctx *gin.Context) {
	page, err := strconv.Atoi(ctx.DefaultQuery("page", "0"))
	if err != nil {
		utils.AbortContextWithError(ctx, err)
		return
	}
	size, err := strconv.Atoi(ctx.DefaultQuery("size", "10"))
	if err != nil {
		utils.AbortContextWithError(ctx, err)
		return
	}
	result, err := c.service.GetAll(ctx, page, size)
	if err != nil {
		utils.AbortContextWithError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, result)
}
func (c *flightController) get(ctx *gin.Context) {
	result, err := c.service.Get(ctx, ctx.Param("flightNumber"))
	if err != nil {
		utils.AbortContextWithError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, result)
}
