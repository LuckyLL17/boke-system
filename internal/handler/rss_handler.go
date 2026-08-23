package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"podcast-platform/api/dto"
	"podcast-platform/internal/repository"
	"podcast-platform/internal/service"
	appErr "podcast-platform/pkg/errors"
)

type RSSHandler struct {
	rssSvc      *service.RSSService
	channelRepo *repository.ChannelRepository
}

func NewRSSHandler(rssSvc *service.RSSService, channelRepo *repository.ChannelRepository) *RSSHandler {
	return &RSSHandler{rssSvc: rssSvc, channelRepo: channelRepo}
}

func (h *RSSHandler) Feed(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	noCache := c.Query("refresh") != ""
	useCache := !noCache
	feed, err := h.rssSvc.GenerateFeed(id, useCache)
	if err != nil {
		ae, _ := appErr.As(err)
		c.JSON(ae.Code, dto.Err(ae.Code, ae.Message))
		return
	}
	c.Header("Content-Type", "application/rss+xml; charset=utf-8")
	c.Header("Cache-Control", "public, max-age=3600")
	c.Header("Last-Modified", time.Now().UTC().Format(http.TimeFormat))
	c.String(http.StatusOK, feed)
}

func (h *RSSHandler) FeedBySlug(c *gin.Context) {
	slug := strings.TrimSuffix(c.Param("slug"), ".xml")
	ch, err := h.channelRepo.GetBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.Err(404, "not found"))
		return
	}
	feed, err := h.rssSvc.GenerateFeed(ch.ID, true)
	if err != nil {
		ae, _ := appErr.As(err)
		c.JSON(ae.Code, dto.Err(ae.Code, ae.Message))
		return
	}
	c.Header("Content-Type", "application/rss+xml; charset=utf-8")
	c.Header("Cache-Control", "public, max-age=3600")
	c.String(http.StatusOK, feed)
}

func (h *RSSHandler) Validate(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	issues, err := h.rssSvc.ValidateFeed(id)
	if err != nil {
		ae, _ := appErr.As(err)
		c.JSON(ae.Code, dto.Err(ae.Code, ae.Message))
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{
		"valid":  len(issues) == 0,
		"issues": issues,
	}))
}

func (h *RSSHandler) SubscribeLinks(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	ch, err := h.channelRepo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.Err(404, "not found"))
		return
	}
	links, err := h.rssSvc.GetSubscribeLinks(id, ch.Slug)
	if err != nil {
		ae, _ := appErr.As(err)
		c.JSON(ae.Code, dto.Err(ae.Code, ae.Message))
		return
	}
	c.JSON(http.StatusOK, dto.OK(links))
}

func (h *RSSHandler) RefreshCache(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	_, err := h.rssSvc.GenerateFeed(id, false)
	h.rssSvc.InvalidateCache(id)
	if err != nil {
		ae, _ := appErr.As(err)
		c.JSON(ae.Code, dto.Err(ae.Code, ae.Message))
		return
	}
	c.JSON(http.StatusOK, dto.OK(nil))
}

func (h *RSSHandler) Unsubscribe(c *gin.Context) {
	token := c.Param("token")
	subscriberSvc, ok := c.Get("subscriber_svc")
	if !ok {
		c.JSON(http.StatusNotFound, dto.Err(404, "handler misconfigured"))
		return
	}
	svc, ok := subscriberSvc.(*service.SubscriberService)
	if !ok {
		c.JSON(http.StatusInternalServerError, dto.Err(500, "type error"))
		return
	}
	if err := svc.Unsubscribe(token); err != nil {
		ae, _ := appErr.As(err)
		c.JSON(ae.Code, dto.Err(ae.Code, ae.Message))
		return
	}
	c.String(http.StatusOK, "Unsubscribed successfully.")
}
