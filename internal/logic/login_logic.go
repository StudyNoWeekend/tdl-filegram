package logic

import (
	"context"

	"tdl-filegram/internal/dto/res"
	"tdl-filegram/pkg/telegram"
)

// LoginLogic 登录业务编排
type LoginLogic struct {
	engine *telegram.Engine
}

func NewLoginLogic(engine *telegram.Engine) *LoginLogic {
	return &LoginLogic{engine: engine}
}

// Status 返回登录状态
func (l *LoginLogic) Status(ctx context.Context) *res.LoginStatusRes {
	ready := l.engine.IsReady()
	authed := ready && l.engine.IsAuthenticated(ctx)
	snap := l.engine.LoginStatus()
	if authed && snap.Status != telegram.LoginStatusSuccess {
		snap.Status = telegram.LoginStatusSuccess
	}
	return &res.LoginStatusRes{
		Ready:         ready,
		Authenticated: authed,
		LoginStatus:   snap.Status,
		QRURL:         snap.QRURL,
		Error:         snap.Error,
	}
}

// StartQR 启动二维码登录
func (l *LoginLogic) StartQR(ctx context.Context) (*res.LoginStatusRes, error) {
	if err := l.engine.StartQRLogin(); err != nil {
		return nil, err
	}
	return l.Status(ctx), nil
}

// Submit2FA 提交两步验证密码
func (l *LoginLogic) Submit2FA(_ context.Context, password string) error {
	return l.engine.Submit2FA(password)
}
