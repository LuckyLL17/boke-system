package dto

type PageQuery struct {
	Page     int    `form:"page" binding:"min=1"`
	PageSize int    `form:"page_size" binding:"min=1,max=100"`
	Keyword  string `form:"keyword"`
}

type ChannelCreateDTO struct {
	Title          string `json:"title" binding:"required,min=1,max=200"`
	Description    string `json:"description" binding:"max=5000"`
	CoverImageURL  string `json:"cover_image_url" binding:"max=500"`
	Language       string `json:"language" binding:"max=20"`
	Copyright      string `json:"copyright" binding:"max=500"`
	Author         string `json:"author" binding:"max=100"`
	Email          string `json:"email" binding:"max=100"`
	CategoryID     uint64 `json:"category_id"`
	Explicit       bool   `json:"explicit"`
	ITunesCategory string `json:"itunes_category" binding:"max=100"`
}

type ChannelUpdateDTO struct {
	Title          *string `json:"title" binding:"omitempty,max=200"`
	Description    *string `json:"description" binding:"omitempty,max=5000"`
	CoverImageURL  *string `json:"cover_image_url" binding:"omitempty,max=500"`
	Language       *string `json:"language" binding:"omitempty,max=20"`
	Copyright      *string `json:"copyright" binding:"omitempty,max=500"`
	Author         *string `json:"author" binding:"omitempty,max=100"`
	Email          *string `json:"email" binding:"omitempty,max=100"`
	CategoryID     *uint64 `json:"category_id"`
	Explicit       *bool   `json:"explicit"`
	CustomDomain   *string `json:"custom_domain" binding:"omitempty,max=200"`
	ITunesCategory *string `json:"itunes_category" binding:"omitempty,max=100"`
}

type ChannelListQuery struct {
	Page     int    `form:"page" binding:"min=1"`
	PageSize int    `form:"page_size" binding:"min=1,max=100"`
	Keyword  string `form:"keyword"`
	Status   *int   `form:"status"`
}

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Total   int64       `json:"total,omitempty"`
	Page    int         `json:"page,omitempty"`
	Pages   int         `json:"pages,omitempty"`
	Error   bool        `json:"error,omitempty"`
}

func OK(data interface{}) Response {
	return Response{Code: 0, Message: "success", Data: data}
}

func PageResult(data interface{}, total int64, page, pageSize int) Response {
	pages := 0
	if pageSize > 0 {
		pages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}
	return Response{Code: 0, Message: "success", Data: data, Total: total, Page: page, Pages: pages}
}

func Err(code int, msg string) Response {
	return Response{Code: code, Message: msg, Error: true}
}
