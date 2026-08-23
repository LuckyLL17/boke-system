package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"podcast-platform/api/dto"
	"podcast-platform/internal/domain"
	"podcast-platform/internal/middleware"
	"podcast-platform/internal/service"
	appErr "podcast-platform/pkg/errors"
)

type EpisodeHandler struct {
	episodeSvc *service.EpisodeService
	audioSvc   *service.AudioService
	commentSvc *service.CommentService
	statsSvc   *service.StatsService
}

func NewEpisodeHandler(episodeSvc *service.EpisodeService, audioSvc *service.AudioService,
	commentSvc *service.CommentService, statsSvc *service.StatsService) *EpisodeHandler {
	return &EpisodeHandler{
		episodeSvc: episodeSvc,
		audioSvc:   audioSvc,
		commentSvc: commentSvc,
		statsSvc:   statsSvc,
	}
}

func (h *EpisodeHandler) parseID(c *gin.Context) uint64 {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	return id
}

func (h *EpisodeHandler) parseChannelID(c *gin.Context) uint64 {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	return id
}

func (h *EpisodeHandler) UploadAudio(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(400, "no file uploaded"))
		return
	}
	result, err := h.audioSvc.UploadAudio(fileHeader)
	if err != nil {
		ae, _ := appErr.As(err)
		c.JSON(ae.Code, dto.Err(ae.Code, ae.Message))
		return
	}
	c.JSON(http.StatusOK, dto.OK(result))
}

func (h *EpisodeHandler) UploadChunk(c *gin.Context) {
	var q dto.UploadChunkDTO
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(400, err.Error()))
		return
	}
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(400, "no file"))
		return
	}
	done, tempDir, err := h.audioSvc.UploadChunked(fileHeader, q.ChunkIndex, q.TotalChunks, q.UploadID)
	if err != nil {
		ae, _ := appErr.As(err)
		c.JSON(ae.Code, dto.Err(ae.Code, ae.Message))
		return
	}
	if done {
		origName := q.OriginalName
		if origName == "" {
			origName = fileHeader.Filename
		}
		result, err := h.audioSvc.MergeChunks(tempDir, origName)
		if err != nil {
			ae, _ := appErr.As(err)
			c.JSON(ae.Code, dto.Err(ae.Code, ae.Message))
			return
		}
		c.JSON(http.StatusOK, dto.OK(gin.H{"merged": true, "result": result}))
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"merged": false, "chunk": q.ChunkIndex}))
}

func (h *EpisodeHandler) Create(c *gin.Context) {
	channelID := h.parseChannelID(c)
	var body dto.EpisodeCreateDTO
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(400, err.Error()))
		return
	}
	userID := middleware.GetUserID(c)
	req := &service.CreateEpisodeRequest{
		Title:         body.Title,
		EpisodeNumber: body.EpisodeNumber,
		SeasonNumber:  body.SeasonNumber,
		Description:   body.Description,
		ShowNotes:     body.ShowNotes,
		CoverImageURL: body.CoverImageURL,
		AudioFileURL:  body.AudioFileURL,
		AudioFileSize: body.AudioFileSize,
		Duration:      body.Duration,
		SampleRate:    body.SampleRate,
		BitRate:       body.BitRate,
		MimeType:      body.MimeType,
		Status:        domain.EpisodeStatus(body.Status),
		ScheduledAt:   body.ScheduledAt,
		Explicit:      body.Explicit,
		Author:        body.Author,
		CategoryID:    body.CategoryID,
		Tags:          body.Tags,
		ChaptersJSON:  body.ChaptersJSON,
	}
	ep, err := h.episodeSvc.Create(channelID, userID, req)
	if err != nil {
		ae, _ := appErr.As(err)
		c.JSON(ae.Code, dto.Err(ae.Code, ae.Message))
		return
	}
	c.JSON(http.StatusOK, dto.OK(ep))
}

func (h *EpisodeHandler) GetByID(c *gin.Context) {
	id := h.parseID(c)
	ep, err := h.episodeSvc.GetByID(id)
	if err != nil {
		ae, _ := appErr.As(err)
		c.JSON(ae.Code, dto.Err(ae.Code, ae.Message))
		return
	}
	c.JSON(http.StatusOK, dto.OK(ep))
}

func (h *EpisodeHandler) Update(c *gin.Context) {
	id := h.parseID(c)
	var body dto.EpisodeUpdateDTO
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(400, err.Error()))
		return
	}
	userID := middleware.GetUserID(c)
	var status *domain.EpisodeStatus
	if body.Status != nil {
		s := domain.EpisodeStatus(*body.Status)
		status = &s
	}
	req := &service.UpdateEpisodeRequest{
		Title:         body.Title,
		Description:   body.Description,
		ShowNotes:     body.ShowNotes,
		CoverImageURL: body.CoverImageURL,
		Status:        status,
		ScheduledAt:   body.ScheduledAt,
		Explicit:      body.Explicit,
		Tags:          body.Tags,
		ChaptersJSON:  body.ChaptersJSON,
	}
	ep, err := h.episodeSvc.Update(id, userID, req)
	if err != nil {
		ae, _ := appErr.As(err)
		c.JSON(ae.Code, dto.Err(ae.Code, ae.Message))
		return
	}
	c.JSON(http.StatusOK, dto.OK(ep))
}

func (h *EpisodeHandler) Delete(c *gin.Context) {
	id := h.parseID(c)
	userID := middleware.GetUserID(c)
	if err := h.episodeSvc.Delete(id, userID); err != nil {
		ae, _ := appErr.As(err)
		c.JSON(ae.Code, dto.Err(ae.Code, ae.Message))
		return
	}
	c.JSON(http.StatusOK, dto.OK(nil))
}

func (h *EpisodeHandler) ListByChannel(c *gin.Context) {
	channelID := h.parseChannelID(c)
	var q dto.EpisodeListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		q.Page = 1
		q.PageSize = 20
	}
	var status *domain.EpisodeStatus
	if q.Status != "" {
		s := domain.EpisodeStatus(q.Status)
		status = &s
	}
	data, total, err := h.episodeSvc.ListByChannel(channelID, q.Page, q.PageSize, status, q.Keyword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(500, err.Error()))
		return
	}
	c.JSON(http.StatusOK, dto.PageResult(data, total, q.Page, q.PageSize))
}

func (h *EpisodeHandler) ListPublished(c *gin.Context) {
	channelID := h.parseChannelID(c)
	limit := 100
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}
	data, err := h.episodeSvc.ListPublishedByChannel(channelID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(500, err.Error()))
		return
	}
	c.JSON(http.StatusOK, dto.OK(data))
}

func (h *EpisodeHandler) Search(c *gin.Context) {
	var q dto.EpisodeSearchQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(400, err.Error()))
		return
	}
	data, total, err := h.episodeSvc.Search(q.Keyword, q.Page, q.PageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(500, err.Error()))
		return
	}
	c.JSON(http.StatusOK, dto.PageResult(data, total, q.Page, q.PageSize))
}

func (h *EpisodeHandler) Like(c *gin.Context) {
	id := h.parseID(c)
	if err := h.episodeSvc.Like(id, 1); err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(500, err.Error()))
		return
	}
	c.JSON(http.StatusOK, dto.OK(nil))
}

func (h *EpisodeHandler) AddComment(c *gin.Context) {
	epIDFromURL := h.parseID(c)
	var body dto.CommentCreateDTO
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(400, err.Error()))
		return
	}
	epID := body.EpisodeID
	if epID == 0 {
		epID = epIDFromURL
	}
	userID := middleware.GetUserID(c)
	channelID := body.ChannelID
	if channelID == 0 {
		channelID = epIDFromURL
	}
	req := &service.CreateCommentRequest{
		EpisodeID:   epID,
		ChannelID:   channelID,
		ParentID:    body.ParentID,
		Type:        domain.CommentType(body.Type),
		Content:     body.Content,
		Rating:      body.Rating,
		IsDanmaku:   body.IsDanmaku,
		DanmakuTime: body.DanmakuTime,
	}
	ip := c.ClientIP()
	ua := c.GetHeader("User-Agent")
	comment, err := h.commentSvc.Create(userID, req, ip, ua)
	if err != nil {
		ae, _ := appErr.As(err)
		c.JSON(ae.Code, dto.Err(ae.Code, ae.Message))
		return
	}
	c.JSON(http.StatusOK, dto.OK(comment))
}

func (h *EpisodeHandler) ListComments(c *gin.Context) {
	id := h.parseID(c)
	var q dto.CommentQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		q.Page = 1
		q.PageSize = 20
	}
	data, total, err := h.commentSvc.ListByEpisode(id, q.Page, q.PageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(500, err.Error()))
		return
	}
	c.JSON(http.StatusOK, dto.PageResult(data, total, q.Page, q.PageSize))
}

func (h *EpisodeHandler) GetDanmaku(c *gin.Context) {
	id := h.parseID(c)
	limit := 200
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil {
			limit = n
		}
	}
	data, err := h.commentSvc.GetDanmaku(id, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(500, err.Error()))
		return
	}
	c.JSON(http.StatusOK, dto.OK(data))
}

func (h *EpisodeHandler) StartPlayback(c *gin.Context) {
	var body dto.PlaybackStartDTO
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(400, err.Error()))
		return
	}
	userID := middleware.GetUserID(c)
	req := service.PlaybackStartRequest{
		EpisodeID: body.EpisodeID,
		UserID:    userID,
		IPAddress: c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
		Source:    body.Source,
	}
	id, err := h.statsSvc.RecordPlaybackStart(req)
	if err != nil {
		ae, _ := appErr.As(err)
		c.JSON(ae.Code, dto.Err(ae.Code, ae.Message))
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"playback_id": id}))
}
