package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"tdl-filegram/enum"
	"tdl-filegram/internal/dto/req"
	"tdl-filegram/internal/logic"
	"tdl-filegram/utils/response"
)

type JobController struct {
	jobLogic *logic.JobLogic
}

func NewJobController(j *logic.JobLogic) *JobController {
	return &JobController{jobLogic: j}
}

// Get 查询单个任务
func (ctl *JobController) Get(c *gin.Context) {
	r, err := ctl.jobLogic.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.Fail(c, enum.ErrNotFound.Code, enum.ErrNotFound.Msg, http.StatusNotFound)
		return
	}
	response.Success(c, r)
}

// List 分页查询任务列表
func (ctl *JobController) List(c *gin.Context) {
	var r req.PaginationReq
	if err := c.ShouldBindQuery(&r); err != nil {
		response.Fail(c, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, http.StatusBadRequest)
		return
	}
	res, err := ctl.jobLogic.List(c.Request.Context(), r)
	if err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, res)
}

// DownloadFile 查看/下载已完成任务的文件
func (ctl *JobController) DownloadFile(c *gin.Context) {
	path, name, err := ctl.jobLogic.GetFile(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.Fail(c, enum.ErrNotFound.Code, err.Error(), http.StatusNotFound)
		return
	}
	// 以 inline 方式返回，浏览器可直接预览（如视频在线播放）
	c.Header("Content-Disposition", "inline; filename=\""+name+"\"")
	c.File(path)
}

// Pause 暂停下载任务
func (ctl *JobController) Pause(c *gin.Context) {
	if err := ctl.jobLogic.Pause(c.Request.Context(), c.Param("id")); err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, nil)
}

// Retry 重试/继续下载任务
func (ctl *JobController) Retry(c *gin.Context) {
	if err := ctl.jobLogic.Retry(c.Request.Context(), c.Param("id")); err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, nil)
}

// Cancel 取消进行中的下载任务
func (ctl *JobController) Cancel(c *gin.Context) {
	if err := ctl.jobLogic.Cancel(c.Request.Context(), c.Param("id")); err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, nil)
}

// Delete 删除单个任务
func (ctl *JobController) Delete(c *gin.Context) {
	deleteFile := c.Query("delete_file") == "true"
	if err := ctl.jobLogic.Delete(c.Request.Context(), c.Param("id"), deleteFile); err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, nil)
}

// BatchDelete 批量删除任务
func (ctl *JobController) BatchDelete(c *gin.Context) {
	var r req.BatchDeleteReq
	if err := c.ShouldBindJSON(&r); err != nil {
		response.Fail(c, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, http.StatusBadRequest)
		return
	}
	if err := ctl.jobLogic.BatchDelete(c.Request.Context(), r.IDs, r.DeleteFile); err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, nil)
}
