package service

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"podcast-platform/internal/domain"
	"podcast-platform/internal/repository"
	appErr "podcast-platform/pkg/errors"
	"podcast-platform/pkg/utils"
)

type SubscriberService struct {
	subscriberRepo *repository.SubscriberRepository
	channelRepo    *repository.ChannelRepository
}

func NewSubscriberService(subscriberRepo *repository.SubscriberRepository, channelRepo *repository.ChannelRepository) *SubscriberService {
	return &SubscriberService{
		subscriberRepo: subscriberRepo,
		channelRepo:    channelRepo,
	}
}

type SubscribeRequest struct {
	Email  string `json:"email"`
	Source string `json:"source"`
}

func (s *SubscriberService) Subscribe(channelID uint64, req *SubscribeRequest, ip, ua string) (*domain.Subscriber, error) {
	if req.Email == "" {
		return nil, appErr.ErrInvalidParams
	}
	ch, err := s.channelRepo.GetByID(channelID)
	if err != nil {
		return nil, appErr.ErrChannelNotFound
	}
	_ = ch
	existing, err := s.subscriberRepo.FindByEmail(channelID, req.Email)
	if err == nil && existing != nil {
		if existing.Status == 1 {
			return existing, nil
		}
		existing.Status = 1
		existing.SubscribedAt = utils.Now()
		existing.UnsubscribedAt = nil
		existing.UnsubToken = generateToken()
		existing.IPAddress = ip
		existing.UserAgent = ua
		if err := s.subscriberRepo.Update(existing); err != nil {
			return nil, appErr.Wrap(err, 500, "resubscribe")
		}
		_ = s.channelRepo.IncrementSubscribers(channelID, 1)
		return existing, nil
	}
	sub := &domain.Subscriber{
		ChannelID:    channelID,
		Email:        req.Email,
		Source:       req.Source,
		IPAddress:    ip,
		UserAgent:    ua,
		Status:       1,
		UnsubToken:   generateToken(),
		SubscribedAt: utils.Now(),
	}
	if err := s.subscriberRepo.Create(sub); err != nil {
		return nil, appErr.Wrap(err, 500, "subscribe")
	}
	_ = s.channelRepo.IncrementSubscribers(channelID, 1)
	return sub, nil
}

func (s *SubscriberService) Unsubscribe(token string) error {
	sub, err := s.subscriberRepo.FindByToken(token)
	if err != nil {
		return appErr.ErrNotFound
	}
	if sub.Status != 1 {
		return nil
	}
	if err := s.subscriberRepo.Unsubscribe(sub.ID); err != nil {
		return appErr.Wrap(err, 500, "unsubscribe")
	}
	_ = s.channelRepo.IncrementSubscribers(sub.ChannelID, -1)
	return nil
}

func (s *SubscriberService) ListByChannel(channelID, ownerID uint64, page, pageSize int, status *int) ([]domain.Subscriber, int64, error) {
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
	if pageSize < 1 || pageSize > 500 {
		pageSize = 50
	}
	return s.subscriberRepo.ListByChannel(channelID, page, pageSize, status)
}

func (s *SubscriberService) CountByChannel(channelID uint64) (int64, error) {
	return s.subscriberRepo.CountByChannel(channelID)
}

func (s *SubscriberService) GrowthByDate(channelID uint64, days int) ([]map[string]interface{}, error) {
	from := utils.DaysAgo(days)
	to := utils.EndOfDay(utils.Now())
	return s.subscriberRepo.GroupByDate(channelID, from, to)
}

func (s *SubscriberService) SourceDistribution(channelID uint64) ([]map[string]interface{}, error) {
	return s.subscriberRepo.GroupBySource(channelID)
}

func (s *SubscriberService) ExportEmails(channelID, ownerID uint64) ([]string, error) {
	ch, err := s.channelRepo.GetByID(channelID)
	if err != nil {
		return nil, appErr.ErrChannelNotFound
	}
	if ch.OwnerID != ownerID {
		return nil, appErr.ErrNoPermission
	}
	return s.subscriberRepo.ExportEmails(channelID)
}

func (s *SubscriberService) Delete(channelID, subscriberID, ownerID uint64) error {
	ch, err := s.channelRepo.GetByID(channelID)
	if err != nil {
		return appErr.ErrChannelNotFound
	}
	if ch.OwnerID != ownerID {
		return appErr.ErrNoPermission
	}
	sub, err := s.subscriberRepo.GetByID(subscriberID)
	if err != nil {
		return appErr.ErrNotFound
	}
	if sub.Status == 1 {
		_ = s.channelRepo.IncrementSubscribers(channelID, -1)
	}
	return s.subscriberRepo.Delete(subscriberID)
}

func generateToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

var _ = time.Now()
