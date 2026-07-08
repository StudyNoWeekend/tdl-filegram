package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"tdl-filegram/enum"
	"tdl-filegram/internal/dto/req"
	"tdl-filegram/internal/logic"
	"tdl-filegram/utils/response"
)

type LoginController struct {
	loginLogic *logic.LoginLogic
}

func NewLoginController(l *logic.LoginLogic) *LoginController {
	return &LoginController{loginLogic: l}
}

// Status 登录状态
func (ctl *LoginController) Status(c *gin.Context) {
	response.Success(c, ctl.loginLogic.Status(c.Request.Context()))
}

// StartQR 启动二维码登录
func (ctl *LoginController) StartQR(c *gin.Context) {
	r, err := ctl.loginLogic.StartQR(c.Request.Context())
	if err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, r)
}

// Submit2FA 提交两步验证密码
func (ctl *LoginController) Submit2FA(c *gin.Context) {
	var r req.Submit2FAReq
	if err := c.ShouldBindJSON(&r); err != nil {
		response.Fail(c, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, http.StatusBadRequest)
		return
	}
	if err := ctl.loginLogic.Submit2FA(c.Request.Context(), r.Password); err != nil {
		response.Fail(c, enum.ErrInvalidParam.Code, err.Error(), http.StatusBadRequest)
		return
	}
	response.Success(c, nil)
}
