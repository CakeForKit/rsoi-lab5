package custom_error

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mdobak/go-xerrors"
)

type HttpError interface {
	Error() string
	Code() int
	ToGinJson() any
}

type baseHttpError struct {
	err  error
	code int
}

func NewHttpError(err string, opts ...HttpErrorOption) HttpError {
	httpError := &baseHttpError{err: xerrors.New(err), code: http.StatusInternalServerError}
	for _, opt := range opts {
		opt(httpError)
	}
	return httpError
}

func (httpErr baseHttpError) Error() string { return xerrors.Sprint(httpErr.err) }

func (httpErr baseHttpError) Code() int { return httpErr.code }

func (httpErr baseHttpError) ToGinJson() any {
	ginJson := gin.H{"message": httpErr.err.Error()}
	return ginJson
}

type HttpErrorOption func(*baseHttpError)

func HttpErrorWithCode(code int) HttpErrorOption {
	return func(httpErr *baseHttpError) {
		httpErr.code = code
	}
}

func IllegalArgumentError(err string) error {
	return NewHttpError(fmt.Sprintf("illegal argument %s", err), HttpErrorWithCode(http.StatusBadRequest))
}

func ParseZeroValueError() error {
	return NewHttpError("parse zero value", HttpErrorWithCode(http.StatusBadRequest))
}

func NotFoundError(param string) error {
	return NewHttpError(fmt.Sprintf("%s not found", param), HttpErrorWithCode(http.StatusNotFound))
}

func InvalidUriError(uri string) error {
	return NewHttpError(fmt.Sprintf("uri %s is not valid", uri))
}
