package api

import (
	"net/http"
	"time"

	"proxy_forwarder/gost/core/auth"
	"proxy_forwarder/gost/core/logger"

	"github.com/gin-gonic/gin"
)

func mwLogger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// start time
		startTime := time.Now()
		// Processing request
		ctx.Next()
		duration := time.Since(startTime)

		logger.Default().WithFields(map[string]any{
			"kind":     "api",
			"method":   ctx.Request.Method,
			"uri":      ctx.Request.RequestURI,
			"code":     ctx.Writer.Status(),
			"client":   ctx.ClientIP(),
			"duration": duration,
		}).Infof("| %3d | %13v | %15s | %-7s %s",
			ctx.Writer.Status(), duration, ctx.ClientIP(), ctx.Request.Method, ctx.Request.RequestURI)
	}
}

func mwBasicAuth(auther auth.Authenticator) gin.HandlerFunc {
	return func(c *gin.Context) {
		if auther == nil {
			return
		}
		u, p, _ := c.Request.BasicAuth()
		if !auther.Authenticate(c, u, p) {
			c.AbortWithStatus(http.StatusUnauthorized)
		}
	}
}
