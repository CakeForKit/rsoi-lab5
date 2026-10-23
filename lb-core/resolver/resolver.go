package resolver

import (
	"reflect"

	"github.com/CakeForKit/rsoi-lab2/lb-core/custom_error"
	"github.com/CakeForKit/rsoi-lab2/lb-core/utils"
	"github.com/gin-gonic/gin"
)

func Resolver[T any](ctx *gin.Context) {
	var obj T
	if err := ctx.ShouldBindJSON(&obj); err != nil {
		utils.AbortContextWithError(ctx, custom_error.IllegalArgumentError(err.Error()))
		return
	}
	if reflect.ValueOf(obj).IsZero() {
		utils.AbortContextWithError(ctx, custom_error.ParseZeroValueError())
		return
	}
	utils.SetRequestBody(ctx, obj)
}
