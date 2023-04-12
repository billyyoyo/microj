package api

import (
	"context"
	"fmt"
	"github.com/billyyoyo/microj/app"
	"github.com/billyyoyo/microj/errs"
	"github.com/billyyoyo/microj/logger"
	"github.com/billyyoyo/microj/util"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"
)

type Api struct {
	Method string
	Path   string
	Func   gin.HandlerFunc
}

type ApiServer struct {
	apis    []Api
	filters []gin.HandlerFunc
	s       *http.Server
}

func (s *ApiServer) Init() {

}

func (s *ApiServer) Run() {
	logger.Info("init http api server")
	gin.SetMode(app.Mode())
	router := gin.New()
	router.Use(apiLogger())
	router.Use(apiRecover())
	if len(s.filters) > 0 {
		for _, f := range s.filters {
			router.Use(f)
		}
	}
	for _, api := range s.apis {
		if strings.ToLower(api.Method) == "get" {
			router.GET(api.Path, api.Func)
		} else if strings.ToLower(api.Method) == "post" {
			router.POST(api.Path, api.Func)
		}
	}
	s.s = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", app.Addr(), app.Port()),
		Handler: router,
	}
	if err := s.s.ListenAndServe(); err != nil {
		logger.Fatal("server startup failed", err)
	}
}

func (s *ApiServer) Stop() {
	s.s.Shutdown(context.Background())
}

func (s *ApiServer) RegFilter(fs []gin.HandlerFunc) *ApiServer {
	s.filters = append(s.filters, fs...)
	return s
}
func (s *ApiServer) RegController(arr []Api) *ApiServer {
	s.apis = append(s.apis, arr...)
	return s
}

func apiLogger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		ctx.Next()
		end := time.Now()
		latencyTime := end.Sub(start)
		reqMethod := ctx.Request.Method
		reqUri := ctx.Request.RequestURI
		statusCode := ctx.Writer.Status()
		clientIP := ctx.ClientIP()
		logger.Infof("code=%d took=%dms ip=%s method=%s path=%s",
			statusCode,
			latencyTime.Milliseconds(),
			clientIP,
			reqMethod,
			reqUri,
		)
	}
}

func apiRecover() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				req, _ := httputil.DumpRequest(ctx.Request, false)
				if ne, ok := err.(errs.MicroError); ok {
					ctx.JSON(http.StatusOK, app.FailedResult(ne.Code(), ne.Error()))
					logger.Error("Recover from panic",
						err.(error),
						logger.Val{K: "request", V: util.Bytes2str(req)},
					)
				} else if ne, ok := err.(error); ok {
					ctx.JSON(http.StatusOK, app.FailedResult(errs.ERRCODE_COMMON, ne.Error()))
					logger.Error("Recover from panic",
						err.(error),
						logger.Val{K: "request", V: util.Bytes2str(req)},
					)
				} else {
					ctx.JSON(http.StatusOK, app.FailedResult(errs.ERRCODE_COMMON, errs.ERRMSG_UNKNOWN))
					logger.Error("Recover from panic",
						nil,
						logger.Val{K: "request", V: util.Bytes2str(req)},
						logger.Val{K: "error", V: err},
					)
				}
			}
		}()
		ctx.Next()
	}
}
