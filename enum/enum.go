package enum

// BizError 业务错误，统一携带错误码、消息、HTTP 状态码
type BizError struct {
	Code     int
	Msg      string
	HttpCode int
}

func (e *BizError) Error() string { return e.Msg }

func NewBizError(code int, msg string, httpCode int) *BizError {
	return &BizError{Code: code, Msg: msg, HttpCode: httpCode}
}

// 全局业务错误码
var (
	ErrInvalidParam     = NewBizError(10040001, "请求参数错误", 400)
	ErrUnauthorized     = NewBizError(10040101, "未登录，请先扫码登录", 401)
	ErrNotFound         = NewBizError(10040401, "资源不存在", 404)
	ErrTelegramNotReady = NewBizError(10050301, "Telegram 未就绪，请稍后", 503)
	ErrDownloadFailed   = NewBizError(10050002, "下载失败", 500)
	ErrInternalServer   = NewBizError(10050001, "系统内部错误", 500)
	ErrFileExists       = NewBizError(10040901, "同名文件已存在，请修改文件名", 409)
)

// 任务状态
const (
	JobStatusPending     = "pending"
	JobStatusDownloading = "downloading"
	JobStatusSuccess     = "success"
	JobStatusFailed      = "failed"
	JobStatusPaused      = "paused"
	JobStatusCancelled   = "cancelled"
)
