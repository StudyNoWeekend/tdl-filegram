package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"tdl-filegram/enum"
	"tdl-filegram/utils/response"
)

// Recovery panic 兜底中间件，捕获 panic 后记录堆栈并返回统一错误
func Recovery(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Error("panic recovered",
					zap.Any("error", r),
					zap.String("stack", string(debug.Stack())),
					zap.String("trace_id", c.GetString("trace_id")),
				)
				response.Fail(c, enum.ErrInternalServer.Code, enum.ErrInternalServer.Msg, http.StatusInternalServerError)
				c.Abort()
			}
		}()
		c.Next()
	}
}
