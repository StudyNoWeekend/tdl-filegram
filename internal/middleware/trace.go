package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Trace 为每个请求注入 trace_id，并回写到响应头
func Trace() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader("X-Trace-Id")
		if traceID == "" {
			traceID = uuid.NewString()
		}
		c.Set("trace_id", traceID)
		c.Header("X-Trace-Id", traceID)
		c.Next()
	}
}
