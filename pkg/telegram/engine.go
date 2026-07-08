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

	runCtx  context.Context
	pool    dcpool.Pool
	manager *peers.Manager

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
		cfg:      cfg,
		log:      log,
		storage:  st,
		client:   client,
		dispatch: dispatch,
	}, nil
}

// Run 启动 telegram client，client 就绪后执行 serve（通常启动 HTTP 服务）
func (e *Engine) Run(ctx context.Context, serve func(ctx context.Context) error) error {
	return e.client.Run(ctx, func(ctx context.Context) error {
		e.runCtx = ctx
		e.pool = dcpool.NewPool(e.client, int64(e.cfg.PoolSize),
			tclient.NewDefaultMiddlewares(ctx, e.cfg.ReconnectTimeout)...)
		e.manager = peers.Options{Storage: storage.NewPeers(e.storage)}.
			Build(e.pool.Default(ctx))
		e.log.Info("telegram engine ready")
		return serve(ctx)
	})
}

// RunCtx 返回 client.Run 的上下文，供 logic 层发起 telegram API 调用
func (e *Engine) RunCtx() context.Context { return e.runCtx }

// IsReady 表示 telegram client 是否已连接就绪
func (e *Engine) IsReady() bool { return e.runCtx != nil }

// Close 关闭底层存储
func (e *Engine) Close() error {
	if c, ok := e.storage.(interface{ Close() error }); ok {
		return c.Close()
	}
	return nil
}

// Storage 返回 session 存储
func (e *Engine) Storage() storage.Storage { return e.storage }
