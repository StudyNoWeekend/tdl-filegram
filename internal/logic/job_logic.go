package logic

import (
	"context"
	"errors"
	"os"
	"time"

	"tdl-filegram/enum"
	"tdl-filegram/internal/dto/req"
	"tdl-filegram/internal/dto/res"
	"tdl-filegram/internal/model"
)

// JobLogic 任务查询业务
type JobLogic struct {
	jobModel *model.JobModel
	download *DownloadLogic
}

func NewJobLogic(jobModel *model.JobModel, download *DownloadLogic) *JobLogic {
	return &JobLogic{jobModel: jobModel, download: download}
}

// Get 查询单个任务（合并实时进度）
func (l *JobLogic) Get(ctx context.Context, id string) (*res.JobRes, error) {
	job, err := l.jobModel.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return l.toJobRes(job), nil
}

// List 分页查询任务列表
func (l *JobLogic) List(ctx context.Context, r req.PaginationReq) (*res.JobListRes, error) {
	page := r.Page
	if page < 1 {
		page = 1
	}
	pageSize := r.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	jobs, total, err := l.jobModel.List(ctx, page, pageSize)
	if err != nil {
		return nil, err
	}
	list := make([]*res.JobRes, 0, len(jobs))
	for _, j := range jobs {
		list = append(list, l.toJobRes(j))
	}
	return &res.JobListRes{List: list, Total: total, Page: page, PageSize: pageSize}, nil
}

// GetFile 获取已完成任务的文件路径，供控制器下载
func (l *JobLogic) GetFile(ctx context.Context, id string) (path, name string, err error) {
	job, err := l.jobModel.GetByID(ctx, id)
	if err != nil {
		return "", "", err
	}
	if job.Status != enum.JobStatusSuccess {
		return "", "", errors.New("file not ready")
	}
	return job.FilePath, job.FileName, nil
}

func (l *JobLogic) toJobRes(j *model.Job) *res.JobRes {
	r := &res.JobRes{
		ID:              j.ID,
		URL:             j.URL,
		Status:          j.Status,
		FileName:        j.FileName,
		FileSize:        j.FileSize,
		DownloadedBytes: j.DownloadedBytes,
		MIME:            j.MIME,
		Error:           j.Error,
		CreatedAt:       j.CreatedAt,
		FilePath:        j.FilePath,
	}
	// 下载中的任务合并内存实时进度
	if j.Status == enum.JobStatusDownloading {
		if lp := l.download.getProgress(j.ID); lp != nil {
			r.FileName = lp.fileName
			r.FileSize = lp.total
			r.DownloadedBytes = lp.downloaded
			r.MIME = lp.mime
			if !lp.startedAt.IsZero() {
				elapsed := time.Since(lp.startedAt).Seconds()
				if elapsed > 0 && lp.downloaded > 0 {
					r.Speed = int64(float64(lp.downloaded) / elapsed)
					if r.Speed > 0 && r.FileSize > r.DownloadedBytes {
						r.EtaSeconds = (r.FileSize - r.DownloadedBytes) / r.Speed
					}
				}
			}
		}
	}
	if r.FileSize > 0 {
		r.Progress = int(r.DownloadedBytes * 100 / r.FileSize)
	}
	return r
}

// Pause 暂停任务
func (l *JobLogic) Pause(ctx context.Context, id string) error {
	return l.download.Pause(id)
}

// Retry 重试/继续任务
func (l *JobLogic) Retry(ctx context.Context, id string) error {
	return l.download.Retry(id)
}

// Cancel 取消进行中的任务
func (l *JobLogic) Cancel(ctx context.Context, id string) error {
	return l.download.Cancel(id)
}

// Delete 删除单个任务，可选删除本地文件
func (l *JobLogic) Delete(ctx context.Context, id string, deleteFile bool) error {
	job, err := l.jobModel.GetByID(ctx, id)
	if err != nil {
		return err
	}
	// 确定文件路径：优先用 DB 中的 FilePath，否则从内存进度获取，最后从 FileName 构建
	filePath := job.FilePath
	if filePath == "" {
		filePath = l.download.FilePath(id) // 下载中任务从 liveProgress 获取
	}
	if filePath == "" {
		filePath = l.download.FilePathByName(job.FileName) // 已取消/失败任务从 FileName 构建
	}
	// 下载中的任务先取消并等待 process goroutine 退出，避免 goroutine 重新 Save 导致任务复活
	if job.Status == enum.JobStatusDownloading {
		_ = l.download.Cancel(id)
		l.download.Wait(id)
	}
	// 清理本地文件和断点续传元数据
	if filePath != "" {
		if deleteFile {
			os.Remove(filePath)
		}
		os.Remove(filePath + ".resume")
		os.Remove(filePath + ".resume.tmp")
	}
	return l.jobModel.Delete(ctx, id)
}

// BatchDelete 批量删除任务，可选删除本地文件
func (l *JobLogic) BatchDelete(ctx context.Context, ids []string, deleteFile bool) error {
	jobs, err := l.jobModel.GetByIDs(ctx, ids)
	if err != nil {
		return err
	}
	for _, job := range jobs {
		if job.Status == enum.JobStatusDownloading {
			_ = l.download.Cancel(job.ID)
		}
	}
	// 等待所有取消的 process goroutine 退出
	for _, job := range jobs {
		if job.Status == enum.JobStatusDownloading {
			l.download.Wait(job.ID)
		}
	}
	for _, job := range jobs {
		// 确定文件路径：优先用 DB 中的 FilePath，否则从内存进度获取，最后从 FileName 构建
		filePath := job.FilePath
		if filePath == "" {
			filePath = l.download.FilePath(job.ID)
		}
		if filePath == "" {
			filePath = l.download.FilePathByName(job.FileName)
		}
		if filePath != "" {
			if deleteFile {
				os.Remove(filePath)
			}
			os.Remove(filePath + ".resume")
			os.Remove(filePath + ".resume.tmp")
		}
	}
	return l.jobModel.DeleteByIDs(ctx, ids)
}
