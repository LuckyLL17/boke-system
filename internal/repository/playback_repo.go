package repository

import (
	"time"

	"gorm.io/gorm"

	"podcast-platform/internal/domain"
)

type PlaybackRepository struct {
	db *gorm.DB
}

func NewPlaybackRepository(db *gorm.DB) *PlaybackRepository {
	return &PlaybackRepository{db: db}
}

func (r *PlaybackRepository) Create(p *domain.Playback) error {
	return r.db.Create(p).Error
}

func (r *PlaybackRepository) BatchCreate(playbacks []domain.Playback) error {
	if len(playbacks) == 0 {
		return nil
	}
	return r.db.Create(&playbacks).Error
}

func (r *PlaybackRepository) Update(p *domain.Playback) error {
	return r.db.Save(p).Error
}

func (r *PlaybackRepository) CountByChannel(channelID uint64, from, to time.Time) (int64, error) {
	var count int64
	q := r.db.Model(&domain.Playback{}).Where("channel_id = ?", channelID)
	if !from.IsZero() {
		q = q.Where("start_at >= ?", from)
	}
	if !to.IsZero() {
		q = q.Where("start_at <= ?", to)
	}
	err := q.Count(&count).Error
	return count, err
}

func (r *PlaybackRepository) CountUniqueListeners(channelID uint64, from, to time.Time) (int64, error) {
	var count int64
	q := r.db.Model(&domain.Playback{}).Where("channel_id = ?", channelID)
	if !from.IsZero() {
		q = q.Where("start_at >= ?", from)
	}
	if !to.IsZero() {
		q = q.Where("start_at <= ?", to)
	}
	err := q.Distinct("ip_address, user_id").Count(&count).Error
	return count, err
}

func (r *PlaybackRepository) CountByEpisode(episodeID uint64, from, to time.Time) (int64, error) {
	var count int64
	q := r.db.Model(&domain.Playback{}).Where("episode_id = ?", episodeID)
	if !from.IsZero() {
		q = q.Where("start_at >= ?", from)
	}
	if !to.IsZero() {
		q = q.Where("start_at <= ?", to)
	}
	err := q.Count(&count).Error
	return count, err
}

func (r *PlaybackRepository) GroupByDate(channelID uint64, from, to time.Time) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	err := r.db.Model(&domain.Playback{}).
		Select("DATE(start_at) as date, COUNT(*) as count, COUNT(DISTINCT ip_address) as listeners").
		Where("channel_id = ? AND start_at >= ? AND start_at <= ?", channelID, from, to).
		Group("DATE(start_at)").Order("date ASC").Find(&results).Error
	return results, err
}

func (r *PlaybackRepository) GroupByHour(channelID uint64, from, to time.Time) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	err := r.db.Model(&domain.Playback{}).
		Select("EXTRACT(HOUR FROM start_at) as hour, COUNT(*) as count").
		Where("channel_id = ? AND start_at >= ? AND start_at <= ?", channelID, from, to).
		Group("EXTRACT(HOUR FROM start_at)").Order("hour ASC").Find(&results).Error
	return results, err
}

func (r *PlaybackRepository) GroupByDayOfWeek(channelID uint64, from, to time.Time) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	err := r.db.Model(&domain.Playback{}).
		Select("EXTRACT(DOW FROM start_at) as dow, COUNT(*) as count").
		Where("channel_id = ? AND start_at >= ? AND start_at <= ?", channelID, from, to).
		Group("EXTRACT(DOW FROM start_at)").Order("dow ASC").Find(&results).Error
	return results, err
}

func (r *PlaybackRepository) GroupByCountry(channelID uint64, from, to time.Time) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	err := r.db.Model(&domain.Playback{}).
		Select("country, COUNT(*) as count, COUNT(DISTINCT ip_address) as listeners").
		Where("channel_id = ? AND start_at >= ? AND start_at <= ?", channelID, from, to).
		Group("country").Order("count DESC").Limit(20).Find(&results).Error
	return results, err
}

func (r *PlaybackRepository) GroupByDevice(channelID uint64, from, to time.Time) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	err := r.db.Model(&domain.Playback{}).
		Select("device_type, os, browser, COUNT(*) as count").
		Where("channel_id = ? AND start_at >= ? AND start_at <= ?", channelID, from, to).
		Group("device_type, os, browser").Order("count DESC").Limit(20).Find(&results).Error
	return results, err
}

func (r *PlaybackRepository) CompletionRate(episodeID uint64) (float64, int64, error) {
	type result struct {
		Total     int64
		Completed int64
		AvgProg   float64
	}
	var res result
	err := r.db.Model(&domain.Playback{}).
		Select("COUNT(*) as total, SUM(CASE WHEN completed THEN 1 ELSE 0 END) as completed, AVG(progress) as avg_prog").
		Where("episode_id = ?", episodeID).Scan(&res).Error
	if err != nil {
		return 0, 0, err
	}
	return res.AvgProg, res.Total, nil
}

func (r *PlaybackRepository) SaveStatsCache(cache *domain.StatsCache) error {
	return r.db.Session(&gorm.Session{SkipDefaultTransaction: true}).Save(cache).Error
}

func (r *PlaybackRepository) GetStatsCache(channelID uint64, date time.Time) (*domain.StatsCache, error) {
	var cache domain.StatsCache
	err := r.db.Where("channel_id = ? AND DATE(date) = DATE(?)", channelID, date).First(&cache).Error
	if err != nil {
		return nil, err
	}
	return &cache, nil
}

func (r *PlaybackRepository) GetRecentByUser(userID, limit int) ([]domain.Playback, error) {
	var playbacks []domain.Playback
	err := r.db.Where("user_id = ?", userID).Order("start_at DESC").Limit(limit).Find(&playbacks).Error
	return playbacks, err
}

type DailySummaryRow struct {
	Date           time.Time `gorm:"column:date"`
	TotalPlaybacks int64     `gorm:"column:total_playbacks"`
	UniqueUsers    int64     `gorm:"column:unique_users"`
	TotalListenSec int64     `gorm:"column:total_listen_sec"`
}

func (r *PlaybackRepository) SummaryByChannel(channelID uint64, from, to time.Time) ([]DailySummaryRow, error) {
	var rows []DailySummaryRow
	err := r.db.Model(&domain.Playback{}).
		Select(`DATE(start_at) as date,
			COUNT(*) as total_playbacks,
			COUNT(DISTINCT COALESCE(NULLIF(user_id, 0), ip_address)) as unique_users,
			COALESCE(SUM(listen_time), 0) as total_listen_sec`).
		Where("channel_id = ? AND start_at >= ? AND start_at <= ?", channelID, from, to).
		Group("DATE(start_at)").Order("date ASC").Scan(&rows).Error
	return rows, err
}
