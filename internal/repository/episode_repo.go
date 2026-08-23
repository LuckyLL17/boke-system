package repository

import (
	"time"

	"gorm.io/gorm"

	"podcast-platform/internal/domain"
)

type EpisodeRepository struct {
	db *gorm.DB
}

func NewEpisodeRepository(db *gorm.DB) *EpisodeRepository {
	return &EpisodeRepository{db: db}
}

func (r *EpisodeRepository) DB() *gorm.DB { return r.db }

func (r *EpisodeRepository) WithTx(tx *gorm.DB) *EpisodeRepository {
	return &EpisodeRepository{db: tx}
}

func (r *EpisodeRepository) Create(ep *domain.Episode) error {
	return r.db.Create(ep).Error
}

func (r *EpisodeRepository) GetByID(id uint64) (*domain.Episode, error) {
	var ep domain.Episode
	err := r.db.Preload("Chapters").Preload("Channel").Where("id = ?", id).First(&ep).Error
	if err != nil {
		return nil, err
	}
	return &ep, nil
}

func (r *EpisodeRepository) Update(ep *domain.Episode) error {
	return r.db.Save(ep).Error
}

func (r *EpisodeRepository) Delete(id uint64) error {
	return r.db.Delete(&domain.Episode{}, id).Error
}

func (r *EpisodeRepository) ListByChannel(channelID uint64, page, pageSize int, status *domain.EpisodeStatus, keyword string) ([]domain.Episode, int64, error) {
	var episodes []domain.Episode
	var total int64
	query := r.db.Model(&domain.Episode{}).Where("channel_id = ?", channelID)
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	if keyword != "" {
		query = query.Where("title LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err := query.Order("COALESCE(published_at, created_at) DESC").
		Offset(offset).Limit(pageSize).Find(&episodes).Error
	return episodes, total, err
}

func (r *EpisodeRepository) ListPublishedByChannel(channelID uint64, limit int) ([]domain.Episode, error) {
	var episodes []domain.Episode
	err := r.db.Preload("Chapters").
		Where("channel_id = ? AND status = ?", channelID, domain.EpisodePublished).
		Order("published_at DESC").Limit(limit).Find(&episodes).Error
	return episodes, err
}

func (r *EpisodeRepository) ExistsBySlug(channelID uint64, slug string) (bool, error) {
	var count int64
	err := r.db.Model(&domain.Episode{}).
		Where("channel_id = ? AND slug = ?", channelID, slug).Count(&count).Error
	return count > 0, err
}

func (r *EpisodeRepository) IncrementPlayCount(episodeID uint64, delta int64) error {
	result := r.db.Model(&domain.Episode{}).Where("id = ?", episodeID).
		UpdateColumn("play_count", gorm.Expr("play_count + ?", delta))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *EpisodeRepository) IncrementCommentCount(episodeID uint64, delta int) error {
	return r.db.Model(&domain.Episode{}).Where("id = ?", episodeID).
		UpdateColumn("comment_count", gorm.Expr("comment_count + ?", delta)).Error
}

func (r *EpisodeRepository) IncrementLikeCount(episodeID uint64, delta int) error {
	return r.db.Model(&domain.Episode{}).Where("id = ?", episodeID).
		UpdateColumn("like_count", gorm.Expr("like_count + ?", delta)).Error
}

func (r *EpisodeRepository) GetTopByChannel(channelID uint64, limit int) ([]domain.Episode, error) {
	var episodes []domain.Episode
	err := r.db.Where("channel_id = ?", channelID).
		Order("play_count DESC").Limit(limit).Find(&episodes).Error
	return episodes, err
}

func (r *EpisodeRepository) GetScheduledEpisodes(before time.Time) ([]domain.Episode, error) {
	var episodes []domain.Episode
	err := r.db.Where("status = ? AND scheduled_at <= ? AND scheduled_at IS NOT NULL",
		domain.EpisodeScheduled, before).Find(&episodes).Error
	return episodes, err
}

func (r *EpisodeRepository) Publish(episodeID uint64) error {
	now := time.Now()
	return r.db.Model(&domain.Episode{}).Where("id = ?", episodeID).
		Updates(map[string]interface{}{
			"status":       domain.EpisodePublished,
			"published_at": &now,
		}).Error
}

func (r *EpisodeRepository) Search(keyword string, page, pageSize int) ([]domain.Episode, int64, error) {
	var episodes []domain.Episode
	var total int64
	query := r.db.Model(&domain.Episode{}).Where("status = ?", domain.EpisodePublished)
	if keyword != "" {
		query = query.Where("title LIKE ? OR description LIKE ? OR tags LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err := query.Preload("Channel").Order("published_at DESC").
		Offset(offset).Limit(pageSize).Find(&episodes).Error
	return episodes, total, err
}

func (r *EpisodeRepository) GetNextEpisodeNumber(channelID uint64) (int, error) {
	var maxNum *int
	err := r.db.Model(&domain.Episode{}).Where("channel_id = ?", channelID).
		Select("MAX(episode_number)").Scan(&maxNum).Error
	if err != nil {
		return 0, err
	}
	if maxNum == nil {
		return 1, nil
	}
	return *maxNum + 1, nil
}

func (r *EpisodeRepository) ListScheduled(before time.Time) ([]domain.Episode, error) {
	return r.GetScheduledEpisodes(before)
}

func (r *EpisodeRepository) UpdateStatus(id uint64, status string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status": status,
	}
	if status == string(domain.EpisodePublished) {
		updates["published_at"] = &now
	}
	return r.db.Model(&domain.Episode{}).Where("id = ?", id).Updates(updates).Error
}
