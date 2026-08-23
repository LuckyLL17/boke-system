package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"podcast-platform/api/dto"
	"podcast-platform/internal/middleware"
	"podcast-platform/internal/service"
	appErr "podcast-platform/pkg/errors"
	"podcast-platform/pkg/logger"
)

type AuthHandler struct {
	authSvc *service.AuthService
}

func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var body dto.RegisterDTO
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(400, "invalid params: "+err.Error()))
		return
	}
	req := &service.RegisterRequest{
		Username: body.Username,
		Email:    body.Email,
		Password: body.Password,
		Nickname: body.Nickname,
	}
	user, err := h.authSvc.Register(req)
	if err != nil {
		ae, ok := appErr.As(err)
		if !ok {
			logger.Errorf("[auth] register unexpected error: username=%s err=%v", body.Username, err)
			c.JSON(http.StatusInternalServerError, dto.Err(500, "registration failed"))
			return
		}
		if ae.Code >= 500 {
			logger.Errorf("[auth] register failed: code=%d username=%s err=%v", ae.Code, body.Username, err)
		}
		c.JSON(ae.Code, dto.Err(ae.Code, ae.Message))
		return
	}
	c.JSON(http.StatusOK, dto.OK(user.ToProfile()))
}

func (h *AuthHandler) Login(c *gin.Context) {
	var body dto.LoginDTO
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(400, "invalid params: "+err.Error()))
		return
	}
	req := &service.LoginRequest{
		Username: body.Username,
		Password: body.Password,
	}
	resp, err := h.authSvc.Login(req)
	if err != nil {
		ae, ok := appErr.As(err)
		if ok {
			c.JSON(ae.Code, dto.Err(ae.Code, ae.Message))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.Err(500, err.Error()))
		return
	}
	c.SetCookie("token", resp.Token, 3600*24, "/", "", false, true)
	c.JSON(http.StatusOK, dto.OK(resp))
}

func (h *AuthHandler) Logout(c *gin.Context) {
	c.SetCookie("token", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, dto.OK(nil))
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID := middleware.GetUserID(c)
	user, err := h.authSvc.GetUserByID(userID)
	if err != nil {
		ae, _ := appErr.As(err)
		c.JSON(ae.Code, dto.Err(ae.Code, ae.Message))
		return
	}
	c.JSON(http.StatusOK, dto.OK(user.ToProfile()))
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var body dto.PasswordChangeDTO
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(400, err.Error()))
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.authSvc.ChangePassword(userID, body.OldPassword, body.NewPassword); err != nil {
		ae, _ := appErr.As(err)
		c.JSON(ae.Code, dto.Err(ae.Code, ae.Message))
		return
	}
	c.JSON(http.StatusOK, dto.OK(nil))
}

func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	var body dto.ProfileUpdateDTO
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(400, err.Error()))
		return
	}
	userID := middleware.GetUserID(c)
	user, err := h.authSvc.UpdateProfile(userID, body.Nickname, body.AvatarURL)
	if err != nil {
		ae, _ := appErr.As(err)
		c.JSON(ae.Code, dto.Err(ae.Code, ae.Message))
		return
	}
	c.JSON(http.StatusOK, dto.OK(user.ToProfile()))
}
