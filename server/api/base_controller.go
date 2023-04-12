package api

import (
	"github.com/billyyoyo/microj/app"
	"github.com/gin-gonic/gin"
	"net/http"
)

type BaseController struct {
}

func (c *BaseController) RetSuccess(ctx *gin.Context, d any) {
	ctx.JSON(http.StatusOK, app.SuccessResult(d))
}

func (c *BaseController) RetFailed(ctx *gin.Context, code int, msg ...string) {
	ctx.JSON(http.StatusOK, app.FailedResult(code, msg...))
}

func (c *BaseController) RetFailedf(ctx *gin.Context, code int, format string, vals ...any) {
	ctx.JSON(http.StatusOK, app.FailedResultf(code, format, vals...))
}
