package model

import (
	"context"
	"time"
)

// Job 下载任务记录表
type Job struct {
	ID              string    `gorm:"primaryKey;type:text" json:"id"`
	URL             string    `gorm:"type:text;not null" json:"url"`
	Status          string    `gorm:"type:varchar(20);index;not null" json:"status"`
	FileName        string    `gorm:"type:text" json:"file_name"`
	FilePath        string    `gorm:"type:text" json:"-"`
	FileSize        int64     `json:"file_size"`
	DownloadedBytes int64     `json:"downloaded_bytes"`
	MIME            string    `gorm:"type:varchar(100)" json:"mime"`
	Error           string    `gorm:"type:text" json:"error,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type JobModel struct{}

func NewJobModel() *JobModel { return &JobModel{} }

func (m *JobModel) Create(ctx context.Context, j *Job) error {
	return DB.WithContext(ctx).Create(j).Error
}

func (m *JobModel) GetByID(ctx context.Context, id string) (*Job, error) {
	var j Job
	if err := DB.WithContext(ctx).Where("id = ?", id).First(&j).Error; err != nil {
		return nil, err
	}
	return &j, nil
}

func (m *JobModel) List(ctx context.Context, page, pageSize int) ([]*Job, int64, error) {
	var jobs []*Job
	var total int64
	db := DB.WithContext(ctx).Model(&Job{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Order("created_at DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&jobs).Error; err != nil {
		return nil, 0, err
	}
	return jobs, total, nil
}

func (m *JobModel) Save(ctx context.Context, j *Job) error {
	return DB.WithContext(ctx).Save(j).Error
}

func (m *JobModel) Delete(ctx context.Context, id string) error {
	return DB.WithContext(ctx).Where("id = ?", id).Delete(&Job{}).Error
}

func (m *JobModel) DeleteByIDs(ctx context.Context, ids []string) error {
	return DB.WithContext(ctx).Where("id IN ?", ids).Delete(&Job{}).Error
}

func (m *JobModel) GetByIDs(ctx context.Context, ids []string) ([]*Job, error) {
	var jobs []*Job
	if err := DB.WithContext(ctx).Where("id IN ?", ids).Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}
