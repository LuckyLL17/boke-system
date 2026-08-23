package errors

import (
	"errors"
	"net/http"
)

type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

var (
	ErrInvalidParams   = New(http.StatusBadRequest, "invalid parameters")
	ErrUnauthorized    = New(http.StatusUnauthorized, "unauthorized")
	ErrForbidden       = New(http.StatusForbidden, "forbidden")
	ErrNotFound        = New(http.StatusNotFound, "resource not found")
	ErrConflict        = New(http.StatusConflict, "resource already exists")
	ErrInternal        = New(http.StatusInternalServerError, "internal server error")
	ErrUserExists      = New(http.StatusConflict, "user already exists")
	ErrUserNotFound    = New(http.StatusNotFound, "user not found")
	ErrInvalidPassword = New(http.StatusUnauthorized, "invalid password")
	ErrTokenExpired    = New(http.StatusUnauthorized, "token expired")
	ErrTokenInvalid    = New(http.StatusUnauthorized, "invalid token")
	ErrChannelNotFound = New(http.StatusNotFound, "channel not found")
	ErrEpisodeNotFound = New(http.StatusNotFound, "episode not found")
	ErrNoPermission    = New(http.StatusForbidden, "no permission")
	ErrUploadFailed    = New(http.StatusBadRequest, "upload failed")
	ErrInvalidAudio    = New(http.StatusBadRequest, "invalid audio file")
)

func New(code int, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

func Wrap(err error, code int, message string) *AppError {
	msg := message
	if err != nil {
		cause := err.Error()
		if cause != "" {
			if msg == "" {
				msg = cause
			} else {
				msg = message + ": " + cause
			}
		}
	}
	return &AppError{
		Code:    code,
		Message: msg,
		Err:     err,
	}
}

func Is(err, target error) bool {
	return errors.Is(err, target)
}

func As(err error) (*AppError, bool) {
	var appErr *AppError
	ok := errors.As(err, &appErr)
	return appErr, ok
}
