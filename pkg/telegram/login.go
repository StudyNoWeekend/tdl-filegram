package telegram

import (
	"context"
	"sync"

	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram/auth/qrlogin"
	"github.com/gotd/td/tgerr"
	"go.uber.org/zap"
)

// 登录状态
const (
	LoginStatusIdle    = ""
	LoginStatusPending = "pending"
	LoginStatusNeed2FA = "need_2fa"
	LoginStatusSuccess = "success"
	LoginStatusError   = "error"
)

// LoginSnapshot 登录状态快照，供前端轮询
type LoginSnapshot struct {
	Status string `json:"status"`
	QRURL  string `json:"qr_url,omitempty"`
	Error  string `json:"error,omitempty"`
}

// loginState 登录流程的共享状态
type loginState struct {
	mu      sync.Mutex
	status  string
	qrURL   string
	errMsg  string
	twoFACh chan string
}

func (s *loginState) snapshot() LoginSnapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	return LoginSnapshot{Status: s.status, QRURL: s.qrURL, Error: s.errMsg}
}

func (s *loginState) setStatus(v string) {
	s.mu.Lock()
	s.status = v
	s.mu.Unlock()
}

func (s *loginState) setQR(url string) {
	s.mu.Lock()
	s.qrURL = url
	s.mu.Unlock()
}

func (s *loginState) setErr(err error) {
	s.mu.Lock()
	s.status = LoginStatusError
	if err != nil {
		s.errMsg = err.Error()
	}
	s.mu.Unlock()
}

// IsAuthenticated 检查当前是否已登录
func (e *Engine) IsAuthenticated(ctx context.Context) bool {
	status, err := e.client.Auth().Status(ctx)
	if err != nil {
		return false
	}
	return status.Authorized
}

// LoginStatus 返回当前登录流程状态快照
func (e *Engine) LoginStatus() LoginSnapshot {
	e.loginMu.Lock()
	st := e.loginState
	e.loginMu.Unlock()
	if st == nil {
		return LoginSnapshot{Status: LoginStatusIdle}
	}
	return st.snapshot()
}

// StartQRLogin 启动二维码登录流程（幂等，已进行中则直接返回）
func (e *Engine) StartQRLogin() error {
	if !e.IsReady() {
		return errors.New("telegram 未就绪，请检查网络或代理配置")
	}
	e.loginMu.Lock()
	defer e.loginMu.Unlock()
	if e.loginState != nil {
		s := e.loginState.snapshot()
		if s.Status == LoginStatusPending || s.Status == LoginStatusNeed2FA {
			return nil
		}
	}
	e.loginState = &loginState{status: LoginStatusPending, twoFACh: make(chan string, 1)}
	go e.doQRLogin(e.loginState)
	return nil
}

// Submit2FA 提交两步验证密码
func (e *Engine) Submit2FA(password string) error {
	e.loginMu.Lock()
	st := e.loginState
	e.loginMu.Unlock()
	if st == nil {
		return errors.New("no login in progress")
	}
	if st.snapshot().Status != LoginStatusNeed2FA {
		return errors.New("2FA not required")
	}
	select {
	case st.twoFACh <- password:
		return nil
	default:
		return errors.New("2FA password already submitted")
	}
}

// doQRLogin 执行二维码登录，token 刷新时更新状态，必要时等待 2FA
func (e *Engine) doQRLogin(st *loginState) {
	_, err := e.client.QR().Auth(e.runCtx, qrlogin.OnLoginToken(e.dispatch),
		func(ctx context.Context, token qrlogin.Token) error {
			st.setQR(token.URL())
			st.setStatus(LoginStatusPending)
			return nil
		})
	if err != nil {
		// 需要两步验证
		if tgerr.Is(err, "SESSION_PASSWORD_NEEDED") {
			st.setStatus(LoginStatusNeed2FA)
			select {
			case pwd := <-st.twoFACh:
				if _, err := e.client.Auth().Password(e.runCtx, pwd); err != nil {
					st.setErr(err)
					e.log.Error("2FA auth failed", zap.Error(err))
					return
				}
			case <-e.runCtx.Done():
				st.setErr(e.runCtx.Err())
				return
			}
		} else {
			st.setErr(err)
			e.log.Error("QR auth failed", zap.Error(err))
			return
		}
	}
	st.setStatus(LoginStatusSuccess)
	e.log.Info("telegram login success")
}
