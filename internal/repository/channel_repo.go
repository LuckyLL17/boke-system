package repository

import (
	"gorm.io/gorm"

	"podcast-platform/internal/domain"
)

type ChannelRepository struct {
	db *gorm.DB
}

func NewChannelRepository(db *gorm.DB) *ChannelRepository {
	return &ChannelRepository{db: db}
}

func (r *ChannelRepository) Create(ch *domain.Channel) error {
	return r.db.Create(ch).Error
}

func (r *ChannelRepository) GetByID(id uint64) (*domain.Channel, error) {
	var ch domain.Channel
	err := r.db.Preload("Owner").Preload("Category").Where("id = ?", id).First(&ch).Error
	if err != nil {
		return nil, err
	}
	return &ch, nil
}

func (r *ChannelRepository) GetBySlug(slug string) (*domain.Channel, error) {
	var ch domain.Channel
	err := r.db.Preload("Owner").Where("slug = ?", slug).First(&ch).Error
	if err != nil {
		return nil, err
	}
	return &ch, nil
}

func (r *ChannelRepository) GetByCustomDomain(customDomain string) (*domain.Channel, error) {
	var ch domain.Channel
	err := r.db.Where("custom_domain = ?", customDomain).First(&ch).Error
	if err != nil {
		return nil, err
	}
	return &ch, nil
}

func (r *ChannelRepository) Update(ch *domain.Channel) error {
	return r.db.Save(ch).Error
}

func (r *ChannelRepository) Delete(id uint64) error {
	return r.db.Delete(&domain.Channel{}, id).Error
}

func (r *ChannelRepository) ListByOwner(ownerID uint64, page, pageSize int, status *domain.ChannelStatus) ([]domain.Channel, int64, error) {
	var channels []domain.Channel
	var total int64
	query := r.db.Model(&domain.Channel{}).Where("owner_id = ?", ownerID)
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err := query.Preload("Category").Order("updated_at DESC").Offset(offset).Limit(pageSize).Find(&channels).Error
	return channels, total, err
}

func (r *ChannelRepository) ListAll(page, pageSize int, keyword string, status *domain.ChannelStatus) ([]domain.Channel, int64, error) {
	var channels []domain.Channel
	var total int64
	query := r.db.Model(&domain.Channel{})
	if keyword != "" {
		query = query.Where("title LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err := query.Preload("Owner").Preload("Category").Order("id DESC").Offset(offset).Limit(pageSize).Find(&channels).Error
	return channels, total, err
}

func (r *ChannelRepository) ExistsBySlug(slug string) (bool, error) {
	var count int64
	err := r.db.Model(&domain.Channel{}).Where("slug = ?", slug).Count(&count).Error
	return count > 0, err
}

func (r *ChannelRepository) IncrementSubscribers(channelID uint64, delta int) error {
	var ch domain.Channel
	if err := r.db.Select("subscribers").Where("id = ?", channelID).First(&ch).Error; err != nil {
		return err
	}
	next := ch.Subscribers + delta
	return r.db.Model(&domain.Channel{}).
		Where("id = ?", channelID).
		Update("subscribers", next).Error
}

func (r *ChannelRepository) IncrementTotalPlays(channelID uint64, delta int64) error {
	return r.db.Model(&domain.Channel{}).Where("id = ?", channelID).
		UpdateColumn("total_plays", gorm.Expr("total_plays + ?", delta)).Error
}

func (r *ChannelRepository) UpdateStatus(id uint64, status domain.ChannelStatus) error {
	return r.db.Model(&domain.Channel{}).Where("id = ?", id).Update("status", status).Error
}

func (r *ChannelRepository) GetAllIDs(limit int) ([]uint64, error) {
	var ids []uint64
	err := r.db.Model(&domain.Channel{}).Where("status = ?", domain.ChannelApproved).
		Pluck("id", &ids).Limit(limit).Error
	return ids, err
}
