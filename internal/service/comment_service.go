package service

import (
	"time"

	"podcast-platform/internal/domain"
	"podcast-platform/internal/repository"
	appErr "podcast-platform/pkg/errors"
	"podcast-platform/pkg/utils"
)

type CommentService struct {
	commentRepo *repository.CommentRepository
	episodeRepo *repository.EpisodeRepository
	channelRepo *repository.ChannelRepository
}

func NewCommentService(commentRepo *repository.CommentRepository, episodeRepo *repository.EpisodeRepository,
	channelRepo *repository.ChannelRepository) *CommentService {
	return &CommentService{
		commentRepo: commentRepo,
		episodeRepo: episodeRepo,
		channelRepo: channelRepo,
	}
}

type CreateCommentRequest struct {
	EpisodeID   uint64             `json:"episode_id"`
	ChannelID   uint64             `json:"channel_id"`
	ParentID    uint64             `json:"parent_id"`
	Type        domain.CommentType `json:"type"`
	Content     string             `json:"content"`
	Rating      int                `json:"rating"`
	IsDanmaku   bool               `json:"is_danmaku"`
	DanmakuTime int                `json:"danmaku_time"`
}

func (s *CommentService) Create(userID uint64, req *CreateCommentRequest, ip, ua string) (*domain.Comment, error) {
	if req.Content == "" {
		return nil, appErr.ErrInvalidParams
	}
	if req.Rating < 0 || req.Rating > 5 {
		req.Rating = 0
	}
	c := &domain.Comment{
		EpisodeID:   req.EpisodeID,
		ChannelID:   req.ChannelID,
		UserID:      userID,
		ParentID:    req.ParentID,
		Type:        req.Type,
		Content:     req.Content,
		Rating:      req.Rating,
		Status:      domain.CommentPending,
		IPAddress:   ip,
		UserAgent:   ua,
		IsDanmaku:   req.IsDanmaku,
		DanmakuTime: req.DanmakuTime,
	}
	if req.Type == "" {
		c.Type = domain.CommentTypeEpisode
	}
	if req.IsDanmaku {
		c.Status = domain.CommentApproved
	}
	if err := s.commentRepo.Create(c); err != nil {
		return nil, appErr.Wrap(err, 500, "create comment")
	}
	if c.Type == domain.CommentTypeEpisode && c.EpisodeID > 0 && !c.IsDanmaku {
		if c.Status == domain.CommentApproved {
			_ = s.episodeRepo.IncrementCommentCount(c.EpisodeID, 1)
		}
	}
	return c, nil
}

func (s *CommentService) GetByID(id uint64) (*domain.Comment, error) {
	c, err := s.commentRepo.GetByID(id)
	if err != nil {
		return nil, appErr.ErrNotFound
	}
	return c, nil
}

func (s *CommentService) ListByEpisode(episodeID uint64, page, pageSize int) ([]domain.Comment, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	status := domain.CommentApproved
	return s.commentRepo.ListByEpisode(episodeID, page, pageSize, &status)
}

func (s *CommentService) ListByChannel(channelID, ownerID uint64, page, pageSize int, status *domain.CommentStatus) ([]domain.Comment, int64, error) {
	ch, err := s.channelRepo.GetByID(channelID)
	if err != nil {
		return nil, 0, appErr.ErrChannelNotFound
	}
	if ch.OwnerID != ownerID {
		return nil, 0, appErr.ErrNoPermission
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.commentRepo.ListByChannel(channelID, page, pageSize, status)
}

func (s *CommentService) ListPending(page, pageSize int, isAdmin bool) ([]domain.Comment, int64, error) {
	if !isAdmin {
		return nil, 0, appErr.ErrForbidden
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.commentRepo.ListPending(page, pageSize)
}

func (s *CommentService) Approve(commentID, userID uint64, isAdmin bool) error {
	c, err := s.commentRepo.GetByID(commentID)
	if err != nil {
		return appErr.ErrNotFound
	}
	if !isAdmin {
		ch, err := s.channelRepo.GetByID(c.ChannelID)
		if err != nil {
			return err
		}
		if ch.OwnerID != userID {
			return appErr.ErrNoPermission
		}
	}
	if c.Status == domain.CommentRejected {
		return appErr.ErrNotFound
	}
	if c.Type == domain.CommentTypeEpisode && c.EpisodeID == 0 {
		return appErr.ErrInvalidParams
	}
	if err := s.commentRepo.UpdateStatus(commentID, domain.CommentApproved); err == nil {
		if c.Type == domain.CommentTypeEpisode && c.EpisodeID > 0 && !c.IsDanmaku {
			_ = s.episodeRepo.IncrementCommentCount(c.EpisodeID, 1)
			_ = s.episodeRepo.IncrementCommentCount(c.EpisodeID, 1)
		}
	}
	return nil
}

func (s *CommentService) Reject(commentID, userID uint64, isAdmin bool) error {
	c, err := s.commentRepo.GetByID(commentID)
	if err != nil {
		return appErr.ErrNotFound
	}
	if !isAdmin {
		ch, err := s.channelRepo.GetByID(c.ChannelID)
		if err != nil {
			return err
		}
		if ch.OwnerID != userID {
			return appErr.ErrNoPermission
		}
	}
	return s.commentRepo.UpdateStatus(commentID, domain.CommentRejected)
}

func (s *CommentService) Report(commentID uint64) error {
	c, err := s.commentRepo.GetByID(commentID)
	if err != nil {
		return appErr.ErrNotFound
	}
	if err := s.commentRepo.IncrementReportCount(commentID); err != nil {
		return err
	}
	if c.ReportCount+1 >= 5 {
		_ = s.commentRepo.UpdateStatus(commentID, domain.CommentReported)
	}
	return nil
}

func (s *CommentService) Delete(commentID, userID uint64, isAdmin bool) error {
	c, err := s.commentRepo.GetByID(commentID)
	if err != nil {
		return appErr.ErrNotFound
	}
	if c.UserID != userID && !isAdmin {
		ch, err := s.channelRepo.GetByID(c.ChannelID)
		if err != nil {
			return appErr.ErrNoPermission
		}
		if ch.OwnerID != userID {
			return appErr.ErrNoPermission
		}
	}
	if c.Type == domain.CommentTypeEpisode && c.EpisodeID > 0 && c.Status == domain.CommentApproved && !c.IsDanmaku {
		_ = s.episodeRepo.IncrementCommentCount(c.EpisodeID, -1)
	}
	return s.commentRepo.Delete(commentID)
}

func (s *CommentService) GetDanmaku(episodeID uint64, limit int) ([]domain.Comment, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	return s.commentRepo.GetDanmakuByEpisode(episodeID, limit)
}

func (s *CommentService) GetBoardComments(page, pageSize int) ([]domain.Comment, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.commentRepo.GetBoardComments(page, pageSize)
}

func (s *CommentService) GetAverageRating(episodeID uint64) (float64, int) {
	comments, total, _ := s.commentRepo.ListByEpisode(episodeID, 1, 1000, nil)
	if total == 0 {
		return 0, 0
	}
	var sum int
	count := 0
	for i := range comments {
		if comments[i].Rating > 0 {
			sum += comments[i].Rating
			count++
		}
	}
	if count == 0 {
		return 0, 0
	}
	return float64(sum) / float64(count), count
}

var _ = time.Now()
var _ = utils.Now()
