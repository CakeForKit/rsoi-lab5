package controller

import "github.com/gin-gonic/gin"

type HttpController interface {
	RegisterHttpController(router *gin.Engine)
}
