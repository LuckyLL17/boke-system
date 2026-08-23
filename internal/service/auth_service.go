package service

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"podcast-platform/config"
	"podcast-platform/internal/domain"
	"podcast-platform/internal/repository"
	appErr "podcast-platform/pkg/errors"
)

type AuthService struct {
	userRepo *repository.UserRepository
}

func NewAuthService(userRepo *repository.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
}

type LoginResponse struct {
	Token string             `json:"token"`
	User  domain.UserProfile `json:"user"`
}

type JWTClaims struct {
	UserID   uint64          `json:"user_id"`
	Username string          `json:"username"`
	Role     domain.UserRole `json:"role"`
	jwt.RegisteredClaims
}

func (s *AuthService) Register(req *RegisterRequest) (*domain.User, error) {
	exists, err := s.userRepo.ExistsByUsername(req.Username)
	if err != nil {
		return nil, appErr.Wrap(err, 500, "db error")
	}
	if exists {
		return nil, appErr.ErrUserExists
	}
	exists, err = s.userRepo.ExistsByEmail(req.Email)
	if err != nil {
		return nil, appErr.Wrap(err, 500, "db error")
	}
	if exists {
		return nil, appErr.New(409, "email already registered")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, appErr.Wrap(err, 500, "hash password failed")
	}
	user := &domain.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hash),
		Nickname:     req.Nickname,
		Role:         domain.RoleUser,
		Status:       1,
	}
	if err := s.userRepo.Create(user); err != nil {
		return nil, appErr.Wrap(err, 500, "create user failed")
	}
	return user, nil
}

func (s *AuthService) Login(req *LoginRequest) (*LoginResponse, error) {
	user, err := s.userRepo.GetByUsername(req.Username)
	if err != nil {
		user, err = s.userRepo.GetByEmail(req.Username)
		if err != nil {
			return nil, appErr.ErrUserNotFound
		}
	}
	if user.Status != 1 {
		return nil, appErr.New(403, "account disabled")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, appErr.ErrInvalidPassword
	}
	token, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}
	return &LoginResponse{
		Token: token,
		User:  user.ToProfile(),
	}, nil
}

func (s *AuthService) generateToken(user *domain.User) (string, error) {
	hours := config.AppConfig.JWT.ExpireHours
	if hours <= 0 {
		hours = 24
	}
	expires := time.Duration(hours) * time.Hour
	now := time.Now()
	claims := JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expires)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "podcast-platform",
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(config.AppConfig.JWT.Secret))
}

func (s *AuthService) ValidateToken(tokenStr string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(config.AppConfig.JWT.Secret), nil
	})
	if err != nil {
		return nil, appErr.ErrTokenInvalid
	}
	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, appErr.ErrTokenInvalid
	}
	if claims.ExpiresAt.Before(time.Now()) {
		return nil, appErr.ErrTokenExpired
	}
	return claims, nil
}

func (s *AuthService) GetUserByID(id uint64) (*domain.User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErr.ErrUserNotFound
		}
		return nil, appErr.Wrap(err, 500, "get user failed")
	}
	return user, nil
}

func (s *AuthService) ChangePassword(userID uint64, oldPwd, newPwd string) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return appErr.ErrUserNotFound
		}
		return appErr.Wrap(err, 500, "get user failed")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPwd)); err != nil {
		return appErr.ErrInvalidPassword
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPwd), bcrypt.DefaultCost)
	if err != nil {
		return appErr.Wrap(err, 500, "hash password failed")
	}
	user.PasswordHash = string(hash)
	if err := s.userRepo.Update(user); err != nil {
		return appErr.Wrap(err, 500, "update password failed")
	}
	return nil
}

func (s *AuthService) UpdateProfile(userID uint64, nickname, avatarURL string) (*domain.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErr.ErrUserNotFound
		}
		return nil, appErr.Wrap(err, 500, "get user failed")
	}
	if nickname != "" {
		user.Nickname = nickname
	}
	if avatarURL != "" {
		user.AvatarURL = avatarURL
	}
	if err := s.userRepo.Update(user); err != nil {
		return nil, appErr.Wrap(err, 500, "update profile failed")
	}
	return user, nil
}
