package service

import (
	"encoding/json"
	"time"

	"podcast-platform/internal/domain"
	"podcast-platform/internal/repository"
	appErr "podcast-platform/pkg/errors"
	"podcast-platform/pkg/utils"
)

type EpisodeService struct {
	episodeRepo *repository.EpisodeRepository
	chapterRepo *repository.ChapterRepository
	channelRepo *repository.ChannelRepository
	audioSvc    *AudioService
}

func NewEpisodeService(episodeRepo *repository.EpisodeRepository, chapterRepo *repository.ChapterRepository,
	channelRepo *repository.ChannelRepository, audioSvc *AudioService) *EpisodeService {
	return &EpisodeService{
		episodeRepo: episodeRepo,
		chapterRepo: chapterRepo,
		channelRepo: channelRepo,
		audioSvc:    audioSvc,
	}
}

type CreateEpisodeRequest struct {
	Title         string               `json:"title"`
	EpisodeNumber int                  `json:"episode_number"`
	SeasonNumber  int                  `json:"season_number"`
	Description   string               `json:"description"`
	ShowNotes     string               `json:"show_notes"`
	CoverImageURL string               `json:"cover_image_url"`
	AudioFileURL  string               `json:"audio_file_url"`
	AudioFileSize int64                `json:"audio_file_size"`
	Duration      int                  `json:"duration"`
	SampleRate    int                  `json:"sample_rate"`
	BitRate       int                  `json:"bit_rate"`
	MimeType      string               `json:"mime_type"`
	Status        domain.EpisodeStatus `json:"status"`
	ScheduledAt   string               `json:"scheduled_at"`
	Explicit      bool                 `json:"explicit"`
	Author        string               `json:"author"`
	CategoryID    uint64               `json:"category_id"`
	Tags          string               `json:"tags"`
	ChaptersJSON  string               `json:"chapters_json"`
	Chapters      []domain.Chapter     `json:"chapters"`
}

type UpdateEpisodeRequest struct {
	Title         string                `json:"title"`
	Description   string                `json:"description"`
	ShowNotes     string                `json:"show_notes"`
	CoverImageURL string                `json:"cover_image_url"`
	Status        *domain.EpisodeStatus `json:"status"`
	ScheduledAt   string                `json:"scheduled_at"`
	Explicit      *bool                 `json:"explicit"`
	Tags          string                `json:"tags"`
	ChaptersJSON  string                `json:"chapters_json"`
	Chapters      []domain.Chapter      `json:"chapters"`
}

func (s *EpisodeService) Create(channelID, ownerID uint64, req *CreateEpisodeRequest) (*domain.Episode, error) {
	ch, err := s.channelRepo.GetByID(channelID)
	if err != nil {
		return nil, appErr.ErrChannelNotFound
	}
	if ch.OwnerID != ownerID {
		return nil, appErr.ErrNoPermission
	}
	if req.Title == "" || req.AudioFileURL == "" {
		return nil, appErr.ErrInvalidParams
	}
	slug := utils.UniqueSlug(req.Title, func(slug string) bool {
		exists, _ := s.episodeRepo.ExistsBySlug(channelID, slug)
		return exists
	})
	epNum := req.EpisodeNumber
	if epNum == 0 {
		epNum, _ = s.episodeRepo.GetNextEpisodeNumber(channelID)
	}
	var epCatID *uint64
	if req.CategoryID > 0 {
		epCatID = &req.CategoryID
	}
	ep := &domain.Episode{
		ChannelID:     channelID,
		Title:         req.Title,
		Slug:          slug,
		EpisodeNumber: epNum,
		SeasonNumber:  req.SeasonNumber,
		Description:   req.Description,
		ShowNotes:     req.ShowNotes,
		CoverImageURL: req.CoverImageURL,
		AudioFileURL:  req.AudioFileURL,
		AudioFileSize: req.AudioFileSize,
		Duration:      req.Duration,
		SampleRate:    req.SampleRate,
		BitRate:       req.BitRate,
		MimeType:      req.MimeType,
		Status:        req.Status,
		Explicit:      req.Explicit,
		Author:        req.Author,
		CategoryID:    epCatID,
		Tags:          req.Tags,
	}
	if ep.Status == "" {
		ep.Status = domain.EpisodeDraft
	}
	if req.ScheduledAt != "" {
		t, err := utils.ParseDateTime(req.ScheduledAt)
		if err == nil {
			ep.ScheduledAt = &t
			if ep.Status == domain.EpisodeDraft {
				ep.Status = domain.EpisodeScheduled
			}
		}
	}
	if ep.Status == domain.EpisodePublished {
		now := utils.Now()
		ep.PublishedAt = &now
	}
	if err := s.episodeRepo.Create(ep); err != nil {
		return nil, appErr.Wrap(err, 500, "create episode failed")
	}
	chapters := req.Chapters
	if len(chapters) == 0 && req.ChaptersJSON != "" {
		var chs []domain.Chapter
		if json.Unmarshal([]byte(req.ChaptersJSON), &chs) == nil {
			chapters = chs
		}
	}
	if len(chapters) > 0 {
		for i := range chapters {
			chapters[i].EpisodeID = ep.ID
			chapters[i].SortOrder = i
		}
		_ = s.chapterRepo.BatchCreate(chapters)
	}
	return ep, nil
}

func (s *EpisodeService) GetByID(id uint64) (*domain.Episode, error) {
	ep, err := s.episodeRepo.GetByID(id)
	if err != nil {
		return nil, appErr.ErrEpisodeNotFound
	}
	return ep, nil
}

func (s *EpisodeService) Update(id, ownerID uint64, req *UpdateEpisodeRequest) (*domain.Episode, error) {
	ep, err := s.episodeRepo.GetByID(id)
	if err != nil {
		return nil, appErr.ErrEpisodeNotFound
	}
	ch, err := s.channelRepo.GetByID(ep.ChannelID)
	if err != nil {
		return nil, appErr.ErrChannelNotFound
	}
	if ch.OwnerID != ownerID {
		return nil, appErr.ErrNoPermission
	}
	if req.Title != "" {
		ep.Title = req.Title
	}
	if req.Description != "" {
		ep.Description = req.Description
	}
	ep.ShowNotes = req.ShowNotes
	if req.CoverImageURL != "" {
		ep.CoverImageURL = req.CoverImageURL
	}
	if req.Status != nil {
		ep.Status = *req.Status
		if *req.Status == domain.EpisodePublished && ep.PublishedAt == nil {
			now := utils.Now()
			ep.PublishedAt = &now
		}
	}
	if req.ScheduledAt != "" {
		if t, err := utils.ParseDateTime(req.ScheduledAt); err == nil {
			ep.ScheduledAt = &t
		}
	}
	if req.Explicit != nil {
		ep.Explicit = *req.Explicit
	}
	if req.Tags != "" {
		ep.Tags = req.Tags
	}
	if err := s.episodeRepo.Update(ep); err != nil {
		return nil, appErr.Wrap(err, 500, "update failed")
	}
	chapters := req.Chapters
	if len(chapters) == 0 && req.ChaptersJSON != "" {
		var chs []domain.Chapter
		if json.Unmarshal([]byte(req.ChaptersJSON), &chs) == nil {
			chapters = chs
		}
	}
	if len(chapters) > 0 || req.ChaptersJSON == "[]" {
		for i := range chapters {
			chapters[i].ID = 0
		}
		_ = s.chapterRepo.ReplaceByEpisode(ep.ID, chapters)
	}
	return ep, nil
}

func (s *EpisodeService) Delete(id, ownerID uint64) error {
	ep, err := s.episodeRepo.GetByID(id)
	if err != nil {
		return appErr.ErrEpisodeNotFound
	}
	ch, err := s.channelRepo.GetByID(ep.ChannelID)
	if err != nil {
		return appErr.ErrChannelNotFound
	}
	if ch.OwnerID != ownerID {
		return appErr.ErrNoPermission
	}
	return s.episodeRepo.Delete(id)
}

func (s *EpisodeService) ListByChannel(channelID uint64, page, pageSize int, status *domain.EpisodeStatus, keyword string) ([]domain.Episode, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.episodeRepo.ListByChannel(channelID, page, pageSize, status, keyword)
}

func (s *EpisodeService) ListPublishedByChannel(channelID uint64, limit int) ([]domain.Episode, error) {
	if limit <= 0 {
		limit = 100
	}
	return s.episodeRepo.ListPublishedByChannel(channelID, limit)
}

func (s *EpisodeService) Search(keyword string, page, pageSize int) ([]domain.Episode, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.episodeRepo.Search(keyword, page, pageSize)
}

func (s *EpisodeService) IncrementPlay(episodeID uint64) error {
	ep, err := s.episodeRepo.GetByID(episodeID)
	if err != nil {
		return err
	}
	_ = s.episodeRepo.IncrementPlayCount(episodeID, 1)
	_ = s.channelRepo.IncrementTotalPlays(ep.ChannelID, 1)
	return nil
}

func (s *EpisodeService) Like(episodeID uint64, delta int) error {
	return s.episodeRepo.IncrementLikeCount(episodeID, delta)
}

func (s *EpisodeService) ProcessScheduledEpisodes() (int, error) {
	episodes, err := s.episodeRepo.GetScheduledEpisodes(time.Now())
	if err != nil {
		return 0, err
	}
	count := 0
	for i := range episodes {
		if err := s.episodeRepo.Publish(episodes[i].ID); err == nil {
			count++
		}
	}
	return count, nil
}
