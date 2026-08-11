package telegram

import (
	"context"
	"sync"
	"time"

	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"
	"go.uber.org/zap"

	"github.com/iyear/tdl/core/dcpool"
	"github.com/iyear/tdl/core/logctx"
	"github.com/iyear/tdl/core/storage"
	"github.com/iyear/tdl/core/tclient"
)

// Config telegram 引擎配置
type Config struct {
	AppID            int
	AppHash          string
	DataDir          string
	Namespace        string
	PoolSize         int
	ReconnectTimeout time.Duration
	Proxy            string
	DownloadDir      string
	Threads          int
	Limit            int
}

// Engine 封装 gotd/tdl 的 telegram 客户端生命周期。
// 所有 telegram 操作必须在 client.Run 上下文内执行，因此 HTTP 服务需在 Run 回调中启动。
type Engine struct {
	cfg     Config
	log     *zap.Logger
	storage storage.Storage
	client  *telegram.Client

	readyMu     sync.RWMutex
	runCtx      context.Context
	pool        dcpool.Pool
	manager     *peers.Manager
	reconnectCh chan struct{}

	loginMu    sync.Mutex
	loginState *loginState
	dispatch   tg.UpdateDispatcher
}

// NewEngine 创建引擎：打开 BoltDB 存储、创建 telegram client（login=false 复用已有 session）
func NewEngine(cfg Config, log *zap.Logger) (*Engine, error) {
	st, err := newBoltStorage(cfg.DataDir, cfg.Namespace)
	if err != nil {
		return nil, errors.Wrap(err, "create bolt storage")
	}
	dispatch := tg.NewUpdateDispatcher()
	ctx := logctx.With(context.Background(), log)
	client, err := tclient.New(ctx, tclient.Options{
		AppID:            cfg.AppID,
		AppHash:          cfg.AppHash,
		Session:          storage.NewSession(st, false),
		Proxy:            cfg.Proxy,
		ReconnectTimeout: cfg.ReconnectTimeout,
		UpdateHandler:    dispatch,
	})
	if err != nil {
		return nil, errors.Wrap(err, "create telegram client")
	}
	return &Engine{
		cfg:         cfg,
		log:         log,
		storage:     st,
		client:      client,
		dispatch:    dispatch,
		reconnectCh: make(chan struct{}, 1),
	}, nil
}

// Run 启动 telegram client，连接断开后等待手动触发重连（懒重连）。
// client 就绪后执行 serve（通常保持运行），连接断开时重置状态并等待 Reconnect() 调用。
func (e *Engine) Run(ctx context.Context, serve func(ctx context.Context) error) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		err := e.client.Run(ctx, func(runCtx context.Context) error {
			pool := dcpool.NewPool(e.client, int64(e.cfg.PoolSize),
				tclient.NewDefaultMiddlewares(runCtx, e.cfg.ReconnectTimeout)...)
			manager := peers.Options{Storage: storage.NewPeers(e.storage)}.
				Build(pool.Default(runCtx))
			e.readyMu.Lock()
			e.runCtx = runCtx
			e.pool = pool
			e.manager = manager
			e.readyMu.Unlock()
			e.log.Info("telegram engine ready")
			return serve(runCtx)
		})
		// 连接断开，重置状态使 IsReady 返回 false
		e.readyMu.Lock()
		e.runCtx = nil
		e.pool = nil
		e.manager = nil
		e.readyMu.Unlock()
		if err := ctx.Err(); err != nil {
			return err
		}
		e.log.Warn("telegram engine disconnected, waiting for reconnect", zap.Error(err))
		select {
		case <-e.reconnectCh:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// Reconnect 触发重连（非阻塞，多次调用安全）。
// 仅在连接已断开时生效，连接正常时调用无副作用。
func (e *Engine) Reconnect() {
	select {
	case e.reconnectCh <- struct{}{}:
	default:
	}
}

// RunCtx 返回 client.Run 的上下文，供 logic 层发起 telegram API 调用
func (e *Engine) RunCtx() context.Context {
	e.readyMu.RLock()
	defer e.readyMu.RUnlock()
	return e.runCtx
}

// IsReady 表示 telegram client 是否已连接就绪
func (e *Engine) IsReady() bool {
	e.readyMu.RLock()
	ctx := e.runCtx
	e.readyMu.RUnlock()
	return ctx != nil && ctx.Err() == nil
}

// getPool 返回当前连接池（线程安全）
func (e *Engine) getPool() dcpool.Pool {
	e.readyMu.RLock()
	defer e.readyMu.RUnlock()
	return e.pool
}

// getManager 返回当前 peers 管理器（线程安全）
func (e *Engine) getManager() *peers.Manager {
	e.readyMu.RLock()
	defer e.readyMu.RUnlock()
	return e.manager
}

// Close 关闭底层存储
func (e *Engine) Close() error {
	if c, ok := e.storage.(interface{ Close() error }); ok {
		return c.Close()
	}
	return nil
}

// Storage 返回 session 存储
func (e *Engine) Storage() storage.Storage { return e.storage }
