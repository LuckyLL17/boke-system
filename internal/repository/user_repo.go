package repository

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"podcast-platform/internal/domain"
)

var (
	ErrDuplicateUser = errors.New("duplicate user")
	ErrUserNotFound  = errors.New("user not found")
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *domain.User) error {
	if err := r.db.Create(user).Error; err != nil {
		if isDuplicateKeyError(err) {
			return fmt.Errorf("%w: %v", ErrDuplicateUser, err)
		}
		return fmt.Errorf("user insert failed: %w", err)
	}
	return nil
}

func (r *UserRepository) GetByID(id uint64) (*domain.User, error) {
	var user domain.User
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: id=%d", ErrUserNotFound, id)
		}
		return nil, fmt.Errorf("get user by id failed: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) GetByUsername(username string) (*domain.User, error) {
	var user domain.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: username=%s", ErrUserNotFound, username)
		}
		return nil, fmt.Errorf("get user by username failed: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) GetByEmail(email string) (*domain.User, error) {
	var user domain.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: email=%s", ErrUserNotFound, email)
		}
		return nil, fmt.Errorf("get user by email failed: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) Update(user *domain.User) error {
	if err := r.db.Save(user).Error; err != nil {
		return fmt.Errorf("update user failed: %w", err)
	}
	return nil
}

func (r *UserRepository) Delete(id uint64) error {
	if err := r.db.Delete(&domain.User{}, id).Error; err != nil {
		return fmt.Errorf("delete user failed: %w", err)
	}
	return nil
}

func (r *UserRepository) ExistsByUsername(username string) (bool, error) {
	var count int64
	if err := r.db.Model(&domain.User{}).Where("username = ?", username).Count(&count).Error; err != nil {
		return false, fmt.Errorf("check username exists failed: %w", err)
	}
	return count > 0, nil
}

func (r *UserRepository) ExistsByEmail(email string) (bool, error) {
	var count int64
	if err := r.db.Model(&domain.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return false, fmt.Errorf("check email exists failed: %w", err)
	}
	return count > 0, nil
}

func (r *UserRepository) List(page, pageSize int, keyword string) ([]domain.User, int64, error) {
	var users []domain.User
	var total int64
	query := r.db.Model(&domain.User{})
	if keyword != "" {
		query = query.Where("username LIKE ? OR email LIKE ? OR nickname LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count users failed: %w", err)
	}
	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("list users failed: %w", err)
	}
	return users, total, nil
}

func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	switch {
	case strings.Contains(msg, "duplicate key"),
		strings.Contains(msg, "UNIQUE constraint failed"),
		strings.Contains(msg, "23505"):
		return true
	}
	return false
}
