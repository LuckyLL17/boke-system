package repository

import (
	"gorm.io/gorm"

	"podcast-platform/internal/domain"
)

type CategoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) Create(c *domain.Category) error {
	return r.db.Create(c).Error
}

func (r *CategoryRepository) GetByID(id uint64) (*domain.Category, error) {
	var c domain.Category
	err := r.db.Where("id = ?", id).First(&c).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CategoryRepository) GetBySlug(slug string) (*domain.Category, error) {
	var c domain.Category
	err := r.db.Where("slug = ?", slug).First(&c).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CategoryRepository) ListAll() ([]domain.Category, error) {
	var categories []domain.Category
	err := r.db.Where("is_active = ?", true).Order("sort_order ASC, id ASC").Find(&categories).Error
	return categories, err
}

func (r *CategoryRepository) ListTree() ([]domain.Category, error) {
	var categories []domain.Category
	err := r.db.Where("parent_id = 0 AND is_active = ?", true).
		Order("sort_order ASC, id ASC").
		Preload("Children").Find(&categories).Error
	return categories, err
}

func (r *CategoryRepository) Update(c *domain.Category) error {
	return r.db.Save(c).Error
}

func (r *CategoryRepository) Delete(id uint64) error {
	return r.db.Delete(&domain.Category{}, id).Error
}
