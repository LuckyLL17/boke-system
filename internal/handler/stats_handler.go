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
	"podcast-platform/pkg/utils"
)

type StatsHandler struct {
	statsSvc   *service.StatsService
	commentSvc *service.CommentService
}

func NewStatsHandler(statsSvc *service.StatsService, commentSvc *service.CommentService) *StatsHandler {
	return &StatsHandler{statsSvc: statsSvc, commentSvc: commentSvc}
}

func (h *StatsHandler) parseID(c *gin.Context) uint64 {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	return id
}

func (h *StatsHandler) ChannelDashboard(c *gin.Context) {
	var q dto.StatsQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		q.Range = "30d"
	}
	q.ChannelID = h.parseID(c)
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, dto.Err(401, "unauthorized"))
		return
	}
	req := service.RangeRequest{
		ChannelID: q.ChannelID,
		Range:     q.Range,
		StartDate: q.StartDate,
		EndDate:   q.EndDate,
	}
	stats, err := h.statsSvc.GetChannelDashboard(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(500, err.Error()))
		return
	}
	for i := range stats.ListenersByDate {
		if stats.ListenersByDate[i].Count == 0 && i < len(stats.PlaysByDate) {
			stats.ListenersByDate[i] = stats.PlaysByDate[i]
		}
	}
	c.JSON(http.StatusOK, dto.OK(stats))
}

func (h *StatsHandler) ExportCSV(c *gin.Context) {
	channelID := h.parseID(c)
	userID := middleware.GetUserID(c)
	_ = userID
	from := utils.DaysAgo(30)
	if s := c.Query("start_date"); s != "" {
		if t, err := utils.ParseDate(s); err == nil {
			from = t
		}
	}
	to := utils.EndOfDay(utils.Now())
	if s := c.Query("end_date"); s != "" {
		if t, err := utils.ParseDate(s); err == nil {
			to = utils.EndOfDay(t)
		}
	}
	_ = h.statsSvc.ExportCSV(channelID, from, to, c.Writer)
}

func (h *StatsHandler) ListChannelComments(c *gin.Context) {
	channelID := h.parseID(c)
	userID := middleware.GetUserID(c)
	var q dto.CommentQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		q.Page = 1
		q.PageSize = 20
	}
	var statusPtr *domain.CommentStatus
	if q.Status != nil {
		s := domain.CommentStatus(*q.Status)
		statusPtr = &s
	}
	data, total, err := h.commentSvc.ListByChannel(channelID, userID, q.Page, q.PageSize, statusPtr)
	if err != nil {
		ae, _ := appErr.As(err)
		c.JSON(ae.Code, dto.Err(ae.Code, ae.Message))
		return
	}
	c.JSON(http.StatusOK, dto.PageResult(data, total, q.Page, q.PageSize))
}

func (h *StatsHandler) ApproveComment(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("comment_id"), 10, 64)
	userID := middleware.GetUserID(c)
	isAdmin := middleware.IsAdmin(c)
	if err := h.commentSvc.Approve(id, userID, isAdmin); err != nil {
		ae, _ := appErr.As(err)
		c.JSON(ae.Code, dto.Err(ae.Code, ae.Message))
		return
	}
	c.JSON(http.StatusOK, dto.OK(nil))
}

func (h *StatsHandler) RejectComment(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("comment_id"), 10, 64)
	userID := middleware.GetUserID(c)
	isAdmin := middleware.IsAdmin(c)
	if err := h.commentSvc.Reject(id, userID, isAdmin); err != nil {
		ae, _ := appErr.As(err)
		c.JSON(ae.Code, dto.Err(ae.Code, ae.Message))
		return
	}
	c.JSON(http.StatusOK, dto.OK(nil))
}

func (h *StatsHandler) ListPendingComments(c *gin.Context) {
	var q dto.PageQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		q.Page = 1
		q.PageSize = 20
	}
	isAdmin := middleware.IsAdmin(c)
	data, total, err := h.commentSvc.ListPending(q.Page, q.PageSize, isAdmin)
	if err != nil {
		ae, _ := appErr.As(err)
		c.JSON(ae.Code, dto.Err(ae.Code, ae.Message))
		return
	}
	c.JSON(http.StatusOK, dto.PageResult(data, total, q.Page, q.PageSize))
}

func (h *StatsHandler) EpisodeCompletion(c *gin.Context) {
	id := h.parseID(c)
	avgProgress, total, err := h.statsSvc.GetEpisodeCompletion(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(500, err.Error()))
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{
		"avg_progress": avgProgress,
		"total_plays":  total,
	}))
}
