package utils

import (
	"context"
	"errors"
	"net/http"

	"github.com/CakeForKit/rsoi-lab2/lb-core/custom_error"
	"github.com/gin-gonic/gin"
	"github.com/mdobak/go-xerrors"
)

const (
	requestBodyKey = "RequestBody"
)

func SetToContext[T any](ctx context.Context, key string, value T) {
	if ginCtx, ok := ConvertContext(ctx); ok {
		ginCtx.Set(key, value)
	}
}

func GetFromContext[T any](ctx context.Context, key string) (T, bool) {
	ginCtx, ok := ConvertContext(ctx)
	if !ok {
		var zeroValue T
		return zeroValue, false
	}

	value, ok := ginCtx.Get(key)
	if !ok {
		var zeroValue T
		return zeroValue, false
	}

	typed, ok := value.(T)
	if !ok {
		var zeroValue T
		return zeroValue, false
	}
	return typed, true
}

func SetRequestBody[T any](ctx context.Context, body T) {
	SetToContext(ctx, requestBodyKey, body)
}

func MustGetRequestBody[T any](ctx context.Context) T {
	result, _ := GetFromContext[T](ctx, requestBodyKey)
	return result
}

func ConvertContext(ctx context.Context) (*gin.Context, bool) {
	cont, ok := ctx.(*gin.Context)
	return cont, ok
}

func AbortContextWithError(ctx *gin.Context, err error) {
	var httpError custom_error.HttpError
	if errors.As(err, &httpError) {
		ctx.AbortWithStatusJSON(httpError.Code(), httpError.ToGinJson())
	} else {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": xerrors.Sprint(err)})
	}
}

func AbortContextWithOK(ctx *gin.Context) {
	ctx.AbortWithStatusJSON(http.StatusOK, gin.H{"message": "OK"})
}
