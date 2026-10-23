package controller

import (
	"context"
	"errors"
	"net/http"
	"net/http/httputil"
	"strconv"
	"sync"

	"github.com/CakeForKit/rsoi-lab2/lb-core/auth"
	"github.com/CakeForKit/rsoi-lab2/lb-core/config"
	coreConroller "github.com/CakeForKit/rsoi-lab2/lb-core/controller"
	"github.com/CakeForKit/rsoi-lab2/lb-core/model"
	"github.com/CakeForKit/rsoi-lab2/ms-gateway/internal/proxy"
	"github.com/CakeForKit/rsoi-lab2/ms-gateway/internal/service"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

var gatewayCntrlr coreConroller.HttpController
var gatewayCntrlrMutex sync.Mutex

type gatewayController struct {
	proxy   *httputil.ReverseProxy
	service *service.GatewayService
}

func GetGatewayController() (coreConroller.HttpController, error) {
	gatewayCntrlrMutex.Lock()
	gatewayCntrlrMutex.Unlock()

	if gatewayCntrlr != nil {
		return gatewayCntrlr, nil
	}
	gatewayCntrlr = &gatewayController{proxy: proxy.NewGatewayProxy(), service: service.New()}
	return gatewayCntrlr, nil
}

func (controller gatewayController) RegisterHttpController(router *gin.Engine) {
	api := router.Group("/api/v1")
	api.POST("/authorize", controller.authorize)
	api.GET("/callback", controller.callback)
	api.GET("/flights", controller.flights)
	api.GET("/privilege", controller.privilege)
	api.POST("/tickets", controller.purchase)
	api.GET("/tickets", controller.tickets)
	api.GET("/tickets/:ticketUid", controller.ticket)
	api.DELETE("/tickets/:ticketUid", controller.cancel)
	api.GET("/me", controller.me)
	router.NoRoute(controller.transfer)
}

func username(ctx *gin.Context) (string, bool) {
	user := auth.Username(ctx)
	if user == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
		return "", false
	}
	return user, true
}

type authorizationRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (controller gatewayController) authorize(ctx *gin.Context) {
	var request authorizationRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "username and password are required"})
		return
	}
	result, status, err := service.Authorize(ctx.Request.Context(), config.CoreConfig.Auth, request.Username, request.Password)
	if err != nil {
		ctx.JSON(status, gin.H{"message": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, result)
}

func (controller gatewayController) callback(ctx *gin.Context) {
	ctx.Status(http.StatusNoContent)
}

func failure(ctx *gin.Context, err error) {
	var unavailable *service.UnavailableError
	if errors.As(err, &unavailable) {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"message": unavailable.Error()})
		return
	}
	ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
}

func (controller gatewayController) flights(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "0"))
	size, _ := strconv.Atoi(ctx.DefaultQuery("size", "10"))
	result, err := controller.service.Flights(ctx.GetHeader("Authorization"), page, size)
	if err != nil {
		failure(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, result)
}

func (controller gatewayController) privilege(ctx *gin.Context) {
	user, ok := username(ctx)
	if !ok {
		return
	}
	result, err := controller.service.GetBonus(ctx.GetHeader("Authorization"), user)
	if err != nil {
		failure(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, result)
}

func (controller gatewayController) purchase(ctx *gin.Context) {
	user, ok := username(ctx)
	if !ok {
		return
	}
	var request model.TicketPurchaseRequest
	if ctx.ShouldBindJSON(&request) != nil {
		failure(ctx, http.ErrNotSupported)
		return
	}
	result, err := controller.service.Purchase(ctx.GetHeader("Authorization"), user, request)
	if err != nil {
		failure(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, result)
}

func (controller gatewayController) tickets(ctx *gin.Context) {
	user, ok := username(ctx)
	if !ok {
		return
	}
	result, err := controller.service.Tickets(ctx.GetHeader("Authorization"), user)
	if err != nil {
		failure(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, result)
}

func (controller gatewayController) ticket(ctx *gin.Context) {
	user, ok := username(ctx)
	if !ok {
		return
	}
	result, err := controller.service.Ticket(ctx.GetHeader("Authorization"), user, ctx.Param("ticketUid"))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"message": "ticket not found"})
		return
	}
	ctx.JSON(http.StatusOK, result)
}

func (controller gatewayController) cancel(ctx *gin.Context) {
	user, ok := username(ctx)
	if !ok {
		return
	}
	if err := controller.service.Cancel(ctx.GetHeader("Authorization"), user, ctx.Param("ticketUid")); err != nil {
		failure(ctx, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (controller gatewayController) me(ctx *gin.Context) {
	user, ok := username(ctx)
	if !ok {
		return
	}
	result, err := controller.service.Me(ctx.GetHeader("Authorization"), user)
	if err != nil {
		failure(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, result)
}

func (controller gatewayController) transfer(ctx *gin.Context) {
	log.WithContext(ctx).Debug("GatewayController: transfer: Start")
	controller.proxy.ServeHTTP(ctx.Writer, ctx.Request.WithContext(context.WithValue(ctx.Request.Context(), gin.ContextKey, ctx)))
	log.WithContext(ctx).Debug("GatewayController: transfer: Success")
}
