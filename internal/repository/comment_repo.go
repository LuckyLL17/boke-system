package repository

import (
	"time"

	"gorm.io/gorm"

	"podcast-platform/internal/domain"
)

type CommentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

func (r *CommentRepository) Create(c *domain.Comment) error {
	return r.db.Create(c).Error
}

func (r *CommentRepository) GetByID(id uint64) (*domain.Comment, error) {
	var c domain.Comment
	err := r.db.Preload("User").Where("id = ?", id).First(&c).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CommentRepository) Update(c *domain.Comment) error {
	return r.db.Save(c).Error
}

func (r *CommentRepository) Delete(id uint64) error {
	return r.db.Delete(&domain.Comment{}, id).Error
}

func (r *CommentRepository) ListByEpisode(episodeID uint64, page, pageSize int, status *domain.CommentStatus) ([]domain.Comment, int64, error) {
	var comments []domain.Comment
	var total int64
	query := r.db.Model(&domain.Comment{}).
		Where("episode_id = ? AND type = ? AND is_danmaku = ?", episodeID, domain.CommentTypeEpisode, false)
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err := query.Preload("User").Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&comments).Error
	return comments, total, err
}

func (r *CommentRepository) ListByChannel(channelID uint64, page, pageSize int, status *domain.CommentStatus) ([]domain.Comment, int64, error) {
	var comments []domain.Comment
	var total int64
	query := r.db.Model(&domain.Comment{}).Where("channel_id = ?", channelID)
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err := query.Preload("User").Preload("Episode").Order("created_at DESC").
		Offset(offset).Limit(pageSize).Find(&comments).Error
	return comments, total, err
}

func (r *CommentRepository) ListPending(page, pageSize int) ([]domain.Comment, int64, error) {
	var comments []domain.Comment
	var total int64
	query := r.db.Model(&domain.Comment{}).Where("status = ?", domain.CommentPending)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err := query.Preload("User").Preload("Episode").Order("created_at ASC").
		Offset(offset).Limit(pageSize).Find(&comments).Error
	return comments, total, err
}

func (r *CommentRepository) GetDanmakuByEpisode(episodeID uint64, limit int) ([]domain.Comment, error) {
	var comments []domain.Comment
	err := r.db.Where("episode_id = ? AND is_danmaku = ? AND status = ?",
		episodeID, true, domain.CommentApproved).
		Order("danmaku_time ASC").Limit(limit).Find(&comments).Error
	return comments, err
}

func (r *CommentRepository) UpdateStatus(id uint64, status domain.CommentStatus) error {
	return r.db.Model(&domain.Comment{}).Where("id = ?", id).
		Where("status <> ?", status).
		Updates(map[string]interface{}{"status": status, "updated_at": time.Now()}).Error
}

func (r *CommentRepository) IncrementReportCount(id uint64) error {
	return r.db.Model(&domain.Comment{}).Where("id = ?", id).
		UpdateColumn("report_count", gorm.Expr("report_count + 1")).Error
}

func (r *CommentRepository) GetBoardComments(page, pageSize int) ([]domain.Comment, int64, error) {
	var comments []domain.Comment
	var total int64
	query := r.db.Model(&domain.Comment{}).
		Where("type = ? AND status = ?", domain.CommentTypeBoard, domain.CommentApproved)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err := query.Preload("User").Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&comments).Error
	return comments, total, err
}
