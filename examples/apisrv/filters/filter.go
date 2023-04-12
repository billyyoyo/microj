package filters

import (
	"github.com/billyyoyo/microj/app"
	"github.com/billyyoyo/microj/examples/gatewaysrv/codes"
	"github.com/gin-gonic/gin"
	"net/http"
)

func Filters() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		AuthFilter(),
		PermissionFilter(),
	}
}

func AuthFilter() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		whitelist := []string{"/login"}

		for _, p := range whitelist {
			if p == ctx.Request.URL.Path {
				ctx.Next()
				return
			}
		}

		token := ctx.GetHeader("Authorization")
		if token == "" {
			ctx.JSON(http.StatusOK, app.FailedResult(codes.ERR_UNAUTHORIZATION, "unauthorization"))
			return
		}
		ctx.Next()
	}
}

func PermissionFilter() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		permissionlist := []string{"/edit"}
		token := ctx.GetHeader("Authorization")
		for _, p := range permissionlist {
			if p == ctx.Request.URL.Path && token == "123" {
				ctx.JSON(http.StatusOK, app.FailedResult(codes.ERR_NO_PERMISSION, "no permission"))
				return
			}
		}
		ctx.Next()
	}
}
