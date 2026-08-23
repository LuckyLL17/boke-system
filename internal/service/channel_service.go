package service

import (
	"podcast-platform/internal/domain"
	"podcast-platform/internal/repository"
	appErr "podcast-platform/pkg/errors"
	"podcast-platform/pkg/utils"
)

type ChannelService struct {
	channelRepo *repository.ChannelRepository
}

func NewChannelService(channelRepo *repository.ChannelRepository) *ChannelService {
	return &ChannelService{channelRepo: channelRepo}
}

type CreateChannelRequest struct {
	Title          string `json:"title"`
	Description    string `json:"description"`
	CoverImageURL  string `json:"cover_image_url"`
	Language       string `json:"language"`
	Copyright      string `json:"copyright"`
	Author         string `json:"author"`
	Email          string `json:"email"`
	CategoryID     uint64 `json:"category_id"`
	Explicit       bool   `json:"explicit"`
	ITunesCategory string `json:"itunes_category"`
}

type UpdateChannelRequest struct {
	Title          *string `json:"title"`
	Description    *string `json:"description"`
	CoverImageURL  *string `json:"cover_image_url"`
	Language       *string `json:"language"`
	Copyright      *string `json:"copyright"`
	Author         *string `json:"author"`
	Email          *string `json:"email"`
	CategoryID     *uint64 `json:"category_id"`
	Explicit       *bool   `json:"explicit"`
	CustomDomain   *string `json:"custom_domain"`
	ITunesCategory *string `json:"itunes_category"`
}

func (s *ChannelService) Create(ownerID uint64, req *CreateChannelRequest) (*domain.Channel, error) {
	if req.Title == "" {
		return nil, appErr.ErrInvalidParams
	}
	slug := utils.UniqueSlug(req.Title, func(candidate string) bool {
		exists, _ := s.channelRepo.ExistsBySlug(candidate)
		return exists
	})
	var catID *uint64
	if req.CategoryID > 0 {
		catID = &req.CategoryID
	}
	ch := &domain.Channel{
		OwnerID:        ownerID,
		Title:          req.Title,
		Slug:           slug,
		Description:    req.Description,
		CoverImageURL:  req.CoverImageURL,
		Language:       orDefault(req.Language, "zh-CN"),
		Copyright:      req.Copyright,
		Author:         req.Author,
		Email:          req.Email,
		CategoryID:     catID,
		Explicit:       req.Explicit,
		ITunesCategory: req.ITunesCategory,
		Status:         domain.ChannelPending,
	}
	if err := s.channelRepo.Create(ch); err != nil {
		return nil, appErr.Wrap(err, 500, "create channel failed")
	}
	return ch, nil
}

func (s *ChannelService) GetByID(id uint64) (*domain.Channel, error) {
	ch, err := s.channelRepo.GetByID(id)
	if err != nil {
		return nil, appErr.ErrChannelNotFound
	}
	return ch, nil
}

func (s *ChannelService) Update(id, ownerID uint64, req *UpdateChannelRequest) (*domain.Channel, error) {
	ch, err := s.channelRepo.GetByID(id)
	if err != nil {
		return nil, appErr.ErrChannelNotFound
	}
	if ch.OwnerID != ownerID {
		return nil, appErr.ErrNoPermission
	}
	if req.Title != nil {
		if *req.Title == "" {
			return nil, appErr.ErrInvalidParams
		}
		ch.Title = *req.Title
	}
	if req.Description != nil {
		ch.Description = *req.Description
	}
	if req.CoverImageURL != nil {
		ch.CoverImageURL = *req.CoverImageURL
	}
	if req.Language != nil {
		ch.Language = *req.Language
	}
	if req.Copyright != nil {
		ch.Copyright = *req.Copyright
	}
	if req.Author != nil {
		ch.Author = *req.Author
	}
	if req.Email != nil {
		ch.Email = *req.Email
	}
	if req.CategoryID != nil {
		if *req.CategoryID > 0 {
			ch.CategoryID = req.CategoryID
		} else {
			ch.CategoryID = nil
		}
	}
	if req.Explicit != nil {
		ch.Explicit = *req.Explicit
	}
	if req.CustomDomain != nil {
		ch.CustomDomain = *req.CustomDomain
	}
	if req.ITunesCategory != nil {
		ch.ITunesCategory = *req.ITunesCategory
	}
	if err := s.channelRepo.Update(ch); err != nil {
		return nil, appErr.Wrap(err, 500, "update failed")
	}
	return ch, nil
}

func (s *ChannelService) Delete(id, ownerID uint64) error {
	ch, err := s.channelRepo.GetByID(id)
	if err != nil {
		return appErr.ErrChannelNotFound
	}
	if ch.OwnerID != ownerID {
		return appErr.ErrNoPermission
	}
	return s.channelRepo.Delete(id)
}

func (s *ChannelService) ListByOwner(ownerID uint64, page, pageSize int, status *domain.ChannelStatus) ([]domain.Channel, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.channelRepo.ListByOwner(ownerID, page, pageSize, status)
}

func (s *ChannelService) ListAll(page, pageSize int, keyword string, status *domain.ChannelStatus) ([]domain.Channel, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.channelRepo.ListAll(page, pageSize, keyword, status)
}

func (s *ChannelService) Approve(id uint64) error {
	return s.channelRepo.UpdateStatus(id, domain.ChannelApproved)
}

func (s *ChannelService) Reject(id uint64) error {
	return s.channelRepo.UpdateStatus(id, domain.ChannelRejected)
}

func (s *ChannelService) CheckOwner(channelID, userID uint64) bool {
	ch, err := s.channelRepo.GetByID(channelID)
	if err != nil {
		return false
	}
	return ch.OwnerID == userID
}

func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
