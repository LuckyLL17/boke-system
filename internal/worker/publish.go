package worker

import (
	"context"

	"podcast-platform/internal/repository"
	"podcast-platform/internal/service"
	"podcast-platform/pkg/logger"
	"podcast-platform/pkg/utils"
)

type PublishHandler struct {
	episodeRepo *repository.EpisodeRepository
	rssSvc      *service.RSSService
	logger      *logger.Logger
}

func NewPublishHandler(epRepo *repository.EpisodeRepository, rssSvc *service.RSSService, log *logger.Logger) *PublishHandler {
	return &PublishHandler{episodeRepo: epRepo, rssSvc: rssSvc, logger: log}
}

func (h *PublishHandler) Name() string { return "publish-scheduled-episodes" }

func (h *PublishHandler) Run(ctx context.Context) error {
	now := utils.Now()
	list, err := h.episodeRepo.ListScheduled(now)
	if err != nil {
		return err
	}
	if len(list) == 0 {
		return nil
	}
	h.logger.Infof("[publish] found %d scheduled episodes to publish", len(list))
	published := 0
	for i := range list {
		ep := &list[i]
		if err := h.episodeRepo.UpdateStatus(ep.ID, "published"); err != nil {
			h.logger.Errorf("[publish] episode %d failed: %v", ep.ID, err)
			continue
		}
		ep.Status = "published"
		go h.rssSvc.GenerateFeed(ep.ChannelID, false)
		published++
	}
	h.logger.Infof("[publish] %d episodes published", published)
	return nil
}
