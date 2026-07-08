package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"tdl-filegram/enum"
	"tdl-filegram/utils/response"
)

// handleErr 统一处理 logic 层返回的错误，区分业务错误与系统错误
func handleErr(c *gin.Context, err error) {
	var bizErr *enum.BizError
	if errors.As(err, &bizErr) {
		response.Fail(c, bizErr.Code, bizErr.Msg, bizErr.HttpCode)
		return
	}
	response.Fail(c, enum.ErrInternalServer.Code, err.Error(), http.StatusInternalServerError)
}
