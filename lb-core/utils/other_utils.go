package utils

import (
	coreController "github.com/CakeForKit/rsoi-lab2/lb-core/controller"
	"github.com/gin-gonic/gin"
	"github.com/xlab/closer"
)

func CheckedError(err error) {
	closer.Checked(func() error { return err }, true)
}

func CheckAndGet[T any](obj T, err error) T {
	CheckedError(err)
	return obj
}

func RegisterController(router *gin.Engine, controllerFunc func() (coreController.HttpController, error)) {
	CheckAndGet(controllerFunc()).RegisterHttpController(router)
}
