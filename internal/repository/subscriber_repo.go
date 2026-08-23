package repository

import (
	"strings"
	"time"

	"gorm.io/gorm"

	"podcast-platform/internal/domain"
)

type SubscriberRepository struct {
	db *gorm.DB
}

func NewSubscriberRepository(db *gorm.DB) *SubscriberRepository {
	return &SubscriberRepository{db: db}
}

func (r *SubscriberRepository) Create(s *domain.Subscriber) error {
	return r.db.Create(s).Error
}

func (r *SubscriberRepository) GetByID(id uint64) (*domain.Subscriber, error) {
	var s domain.Subscriber
	err := r.db.Where("id = ?", id).First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SubscriberRepository) Update(s *domain.Subscriber) error {
	return r.db.Save(s).Error
}

func (r *SubscriberRepository) Delete(id uint64) error {
	return r.db.Delete(&domain.Subscriber{}, id).Error
}

func (r *SubscriberRepository) ListByChannel(channelID uint64, page, pageSize int, status *int) ([]domain.Subscriber, int64, error) {
	var subscribers []domain.Subscriber
	var total int64
	query := r.db.Model(&domain.Subscriber{}).Where("channel_id = ?", channelID)
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err := query.Order("subscribed_at DESC").Offset(offset).Limit(pageSize).Find(&subscribers).Error
	return subscribers, total, err
}

func (r *SubscriberRepository) CountByChannel(channelID uint64) (int64, error) {
	var count int64
	err := r.db.Model(&domain.Subscriber{}).Where("channel_id = ? AND status = 1", channelID).Count(&count).Error
	return count, err
}

func (r *SubscriberRepository) CountByChannelDate(channelID uint64, from, to time.Time) (int64, error) {
	var count int64
	err := r.db.Model(&domain.Subscriber{}).
		Where("channel_id = ? AND status = 1 AND subscribed_at >= ? AND subscribed_at <= ?",
			channelID, from, to).Count(&count).Error
	return count, err
}

func (r *SubscriberRepository) FindByEmail(channelID uint64, email string) (*domain.Subscriber, error) {
	var s domain.Subscriber
	err := r.db.Where("channel_id = ? AND email = ?", channelID, strings.ToLower(email)).First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SubscriberRepository) FindByToken(token string) (*domain.Subscriber, error) {
	var s domain.Subscriber
	err := r.db.Where("unsub_token = ?", token).First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SubscriberRepository) Unsubscribe(id uint64) error {
	now := time.Now()
	return r.db.Model(&domain.Subscriber{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":          0,
			"unsubscribed_at": &now,
		}).Error
}

func (r *SubscriberRepository) GroupByDate(channelID uint64, from, to time.Time) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	err := r.db.Model(&domain.Subscriber{}).
		Select("DATE(subscribed_at) as date, COUNT(*) as count").
		Where("channel_id = ? AND status = 1 AND subscribed_at >= ? AND subscribed_at <= ?",
			channelID, from, to).
		Group("DATE(subscribed_at)").Order("date ASC").Find(&results).Error
	return results, err
}

func (r *SubscriberRepository) GroupBySource(channelID uint64) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	err := r.db.Model(&domain.Subscriber{}).
		Select("source, COUNT(*) as count").
		Where("channel_id = ? AND status = 1", channelID).
		Group("source").Order("count DESC").Find(&results).Error
	return results, err
}

func (r *SubscriberRepository) ExportEmails(channelID uint64) ([]string, error) {
	var emails []string
	err := r.db.Model(&domain.Subscriber{}).
		Where("channel_id = ? AND status = 1 AND email != ''", channelID).
		Pluck("email", &emails).Error
	return emails, err
}
