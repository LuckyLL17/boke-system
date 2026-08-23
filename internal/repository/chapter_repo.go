package repository

import (
	"gorm.io/gorm"

	"podcast-platform/internal/domain"
)

type ChapterRepository struct {
	db *gorm.DB
}

func NewChapterRepository(db *gorm.DB) *ChapterRepository {
	return &ChapterRepository{db: db}
}

func (r *ChapterRepository) Create(ch *domain.Chapter) error {
	return r.db.Create(ch).Error
}

func (r *ChapterRepository) BatchCreate(chapters []domain.Chapter) error {
	if len(chapters) == 0 {
		return nil
	}
	return r.db.Create(&chapters).Error
}

func (r *ChapterRepository) GetByID(id uint64) (*domain.Chapter, error) {
	var ch domain.Chapter
	err := r.db.Where("id = ?", id).First(&ch).Error
	if err != nil {
		return nil, err
	}
	return &ch, nil
}

func (r *ChapterRepository) ListByEpisode(episodeID uint64) ([]domain.Chapter, error) {
	var chapters []domain.Chapter
	err := r.db.Where("episode_id = ?", episodeID).Order("sort_order ASC, start_time ASC").Find(&chapters).Error
	return chapters, err
}

func (r *ChapterRepository) Update(ch *domain.Chapter) error {
	return r.db.Save(ch).Error
}

func (r *ChapterRepository) Delete(id uint64) error {
	return r.db.Delete(&domain.Chapter{}, id).Error
}

func (r *ChapterRepository) DeleteByEpisode(episodeID uint64) error {
	return r.db.Where("episode_id = ?", episodeID).Delete(&domain.Chapter{}).Error
}

// ReplaceByEpisode replaces all chapters for an episode atomically: existing
// rows are deleted and the supplied chapters (if any) are inserted in the same
// transaction. An empty slice is a legitimate value that clears all chapters.
func (r *ChapterRepository) ReplaceByEpisode(episodeID uint64, chapters []domain.Chapter) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return r.ReplaceByEpisodeTx(tx, episodeID, chapters)
	})
}

// ReplaceByEpisodeTx performs the same replacement as ReplaceByEpisode but on
// an existing transaction/connection handle, so callers can fold the chapter
// rewrite into a larger atomic unit of work.
func (r *ChapterRepository) ReplaceByEpisodeTx(tx *gorm.DB, episodeID uint64, chapters []domain.Chapter) error {
	if tx == nil {
		tx = r.db
	}
	if err := tx.Where("episode_id = ?", episodeID).Delete(&domain.Chapter{}).Error; err != nil {
		return err
	}
	if len(chapters) == 0 {
		return nil
	}
	for i := range chapters {
		chapters[i].ID = 0
		chapters[i].EpisodeID = episodeID
		chapters[i].SortOrder = i
	}
	if err := tx.Create(&chapters).Error; err != nil {
		return err
	}
	return nil
}
