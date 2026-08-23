package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"podcast-platform/api/dto"
	"podcast-platform/internal/domain"
	"podcast-platform/internal/middleware"
	"podcast-platform/internal/service"
	appErr "podcast-platform/pkg/errors"
)

type ChannelHandler struct {
	channelSvc    *service.ChannelService
	subscriberSvc *service.SubscriberService
}

func NewChannelHandler(channelSvc *service.ChannelService, subscriberSvc *service.SubscriberService) *ChannelHandler {
	return &ChannelHandler{channelSvc: channelSvc, subscriberSvc: subscriberSvc}
}

func (h *ChannelHandler) parseID(c *gin.Context) uint64 {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	return id
}

func (h *ChannelHandler) Create(c *gin.Context) {
	var body dto.ChannelCreateDTO
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(400, err.Error()))
		return
	}
	userID := middleware.GetUserID(c)
	req := &service.CreateChannelRequest{
		Title:          body.Title,
		Description:    body.Description,
		CoverImageURL:  body.CoverImageURL,
		Language:       body.Language,
		Copyright:      body.Copyright,
		Author:         body.Author,
		Email:          body.Email,
		CategoryID:     body.CategoryID,
		Explicit:       body.Explicit,
		ITunesCategory: body.ITunesCategory,
	}
	ch, err := h.channelSvc.Create(userID, req)
	if err != nil {
		ae, _ := appErr.As(err)
		c.JSON(ae.Code, dto.Err(ae.Code, ae.Message))
		return
	}
	c.JSON(http.StatusOK, dto.OK(ch))
}

func (h *ChannelHandler) GetByID(c *gin.Context) {
	id := h.parseID(c)
	ch, err := h.channelSvc.GetByID(id)
	if err != nil {
		ae, _ := appErr.As(err)
		c.JSON(ae.Code, dto.Err(ae.Code, ae.Message))
		return
	}
	c.JSON(http.StatusOK, dto.OK(ch))
}

func (h *ChannelHandler) Update(c *gin.Context) {
	id := h.parseID(c)
	var body dto.ChannelUpdateDTO
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(400, err.Error()))
		return
	}
	userID := middleware.GetUserID(c)
	req := &service.UpdateChannelRequest{
		Title:          body.Title,
		Description:    body.Description,
		CoverImageURL:  body.CoverImageURL,
		Language:       body.Language,
		Copyright:      body.Copyright,
		Author:         body.Author,
		Email:          body.Email,
		CategoryID:     body.CategoryID,
		Explicit:       body.Explicit,
		CustomDomain:   body.CustomDomain,
		ITunesCategory: body.ITunesCategory,
	}
	ch, err := h.channelSvc.Update(id, userID, req)
	if err != nil {
		ae, _ := appErr.As(err)
		c.JSON(ae.Code, dto.Err(ae.Code, ae.Message))
		return
	}
	c.JSON(http.StatusOK, dto.OK(ch))
}

func (h *ChannelHandler) Delete(c *gin.Context) {
	id := h.parseID(c)
	userID := middleware.GetUserID(c)
	if err := h.channelSvc.Delete(id, userID); err != nil {
		ae, _ := appErr.As(err)
		c.JSON(ae.Code, dto.Err(ae.Code, ae.Message))
		return
	}
	c.JSON(http.StatusOK, dto.OK(nil))
}

func (h *ChannelHandler) ListMyChannels(c *gin.Context) {
	var q dto.ChannelListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		q.Page = 1
		q.PageSize = 20
	}
	userID := middleware.GetUserID(c)
	var status *domain.ChannelStatus
	if q.Status != nil {
		s := domain.ChannelStatus(*q.Status)
		status = &s
	}
	data, total, err := h.channelSvc.ListByOwner(userID, q.Page, q.PageSize, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(500, err.Error()))
		return
	}
	c.JSON(http.StatusOK, dto.PageResult(data, total, q.Page, q.PageSize))
}

func (h *ChannelHandler) ListAll(c *gin.Context) {
	var q dto.ChannelListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		q.Page = 1
		q.PageSize = 20
	}
	var status *domain.ChannelStatus
	if q.Status != nil {
		s := domain.ChannelStatus(*q.Status)
		status = &s
	}
	data, total, err := h.channelSvc.ListAll(q.Page, q.PageSize, q.Keyword, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(500, err.Error()))
		return
	}
	c.JSON(http.StatusOK, dto.PageResult(data, total, q.Page, q.PageSize))
}

func (h *ChannelHandler) Approve(c *gin.Context) {
	id := h.parseID(c)
	if err := h.channelSvc.Approve(id); err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(500, err.Error()))
		return
	}
	c.JSON(http.StatusOK, dto.OK(nil))
}

func (h *ChannelHandler) Reject(c *gin.Context) {
	id := h.parseID(c)
	if err := h.channelSvc.Reject(id); err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(500, err.Error()))
		return
	}
	c.JSON(http.StatusOK, dto.OK(nil))
}

func (h *ChannelHandler) Subscribe(c *gin.Context) {
	id := h.parseID(c)
	var body dto.SubscribeDTO
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(400, err.Error()))
		return
	}
	source := strings.TrimSpace(body.Source)
	if source == "" {
		source = "unknown"
	}
	req := &service.SubscribeRequest{Email: body.Email, Source: source}
	ip := c.ClientIP()
	ua := c.GetHeader("User-Agent")
	sub, err := h.subscriberSvc.Subscribe(id, req, ip, ua)
	if err != nil {
		ae, _ := appErr.As(err)
		c.JSON(ae.Code, dto.Err(ae.Code, ae.Message))
		return
	}
	c.JSON(http.StatusOK, dto.OK(sub))
}

func (h *ChannelHandler) ListSubscribers(c *gin.Context) {
	id := h.parseID(c)
	var q dto.PageQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		q.Page = 1
		q.PageSize = 50
	}
	userID := middleware.GetUserID(c)
	var status *int
	data, total, err := h.subscriberSvc.ListByChannel(id, userID, q.Page, q.PageSize, status)
	if err != nil {
		ae, _ := appErr.As(err)
		c.JSON(ae.Code, dto.Err(ae.Code, ae.Message))
		return
	}
	c.JSON(http.StatusOK, dto.PageResult(data, total, q.Page, q.PageSize))
}

func (h *ChannelHandler) ExportSubscribers(c *gin.Context) {
	id := h.parseID(c)
	userID := middleware.GetUserID(c)
	emails, err := h.subscriberSvc.ExportEmails(id, userID)
	if err != nil {
		ae, _ := appErr.As(err)
		c.JSON(ae.Code, dto.Err(ae.Code, ae.Message))
		return
	}
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=subscribers.csv")
	c.String(http.StatusOK, "email\n"+joinEmails(emails))
}

func joinEmails(list []string) string {
	result := ""
	for i, e := range list {
		if i > 0 {
			result += "\n"
		}
		result += e
	}
	return result
}
