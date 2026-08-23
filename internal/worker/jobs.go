package worker

import (
	"context"

	"podcast-platform/internal/repository"
	"podcast-platform/internal/service"
	"podcast-platform/pkg/logger"
	"podcast-platform/pkg/utils"
)

type AggregateHandler struct {
	playbackRepo *repository.PlaybackRepository
	channelRepo  *repository.ChannelRepository
	statsSvc     *service.StatsService
	logger       *logger.Logger
}

func NewAggregateHandler(pRepo *repository.PlaybackRepository, cRepo *repository.ChannelRepository,
	statsSvc *service.StatsService, log *logger.Logger) *AggregateHandler {
	return &AggregateHandler{
		playbackRepo: pRepo,
		channelRepo:  cRepo,
		statsSvc:     statsSvc,
		logger:       log,
	}
}

func (h *AggregateHandler) Name() string { return "aggregate-stats" }

func (h *AggregateHandler) Run(ctx context.Context) error {
	yesterday := utils.AddDays(utils.Today(), -1)
	channels, _, err := h.channelRepo.ListAll(1, 10000, "", nil)
	if err != nil {
		return err
	}
	total := 0
	for _, ch := range channels {
		rows, err := h.playbackRepo.SummaryByChannel(ch.ID, yesterday, utils.EndOfDay(yesterday))
		if err != nil {
			h.logger.Errorf("[aggregate] channel %d failed: %v", ch.ID, err)
			continue
		}
		for _, r := range rows {
			if r.TotalPlaybacks == 0 {
				continue
			}
			h.statsSvc.UpdateDailyCache(ch.ID, yesterday, r)
			total++
		}
	}
	h.logger.Infof("[aggregate] %d daily rows updated", total)
	return nil
}

type FeedRefreshHandler struct {
	channelRepo *repository.ChannelRepository
	rssSvc      *service.RSSService
	logger      *logger.Logger
}

func NewFeedRefreshHandler(cRepo *repository.ChannelRepository, rssSvc *service.RSSService, log *logger.Logger) *FeedRefreshHandler {
	return &FeedRefreshHandler{channelRepo: cRepo, rssSvc: rssSvc, logger: log}
}

func (h *FeedRefreshHandler) Name() string { return "refresh-rss-cache" }

func (h *FeedRefreshHandler) Run(ctx context.Context) error {
	ids, err := h.channelRepo.GetAllIDs(10000)
	if err != nil {
		return err
	}
	refreshed := 0
	for _, id := range ids {
		if _, err := h.rssSvc.RefreshCache(id); err != nil {
			h.logger.Errorf("[rss-refresh] channel %d failed: %v", id, err)
			continue
		}
		refreshed++
	}
	h.logger.Infof("[rss-refresh] %d feeds refreshed", refreshed)
	return nil
}
