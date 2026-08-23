package dto

type EpisodeCreateDTO struct {
	Title         string  `json:"title" binding:"required,min=1,max=300"`
	EpisodeNumber int     `json:"episode_number"`
	SeasonNumber  int     `json:"season_number"`
	Description   string  `json:"description" binding:"max=10000"`
	ShowNotes     string  `json:"show_notes" binding:"max=20000"`
	CoverImageURL string  `json:"cover_image_url" binding:"max=500"`
	AudioFileURL  string  `json:"audio_file_url" binding:"required,max=500"`
	AudioFileSize int64   `json:"audio_file_size"`
	Duration      int     `json:"duration"`
	SampleRate    int     `json:"sample_rate"`
	BitRate       int     `json:"bit_rate"`
	MimeType      string  `json:"mime_type" binding:"max=50"`
	Status        string  `json:"status"`
	ScheduledAt   string  `json:"scheduled_at"`
	Explicit      bool    `json:"explicit"`
	Author        string  `json:"author" binding:"max=100"`
	CategoryID    uint64  `json:"category_id"`
	Tags          string  `json:"tags" binding:"max=500"`
	ChaptersJSON  string  `json:"chapters_json"`
}

type EpisodeUpdateDTO struct {
	Title         string `json:"title" binding:"max=300"`
	Description   string `json:"description" binding:"max=10000"`
	ShowNotes     string `json:"show_notes" binding:"max=20000"`
	CoverImageURL string `json:"cover_image_url" binding:"max=500"`
	Status        *string `json:"status"`
	ScheduledAt   string `json:"scheduled_at"`
	Explicit      *bool  `json:"explicit"`
	Tags          string `json:"tags" binding:"max=500"`
	ChaptersJSON  string `json:"chapters_json"`
}

type EpisodeListQuery struct {
	Page     int    `form:"page" binding:"min=1"`
	PageSize int    `form:"page_size" binding:"min=1,max=100"`
	Status   string `form:"status"`
	Keyword  string `form:"keyword"`
}

type EpisodeSearchQuery struct {
	Page     int    `form:"page" binding:"min=1"`
	PageSize int    `form:"page_size" binding:"min=1,max=100"`
	Keyword  string `form:"keyword" binding:"required"`
}

type UploadChunkDTO struct {
	UploadID     string `form:"upload_id" binding:"required"`
	ChunkIndex   int    `form:"chunk_index" binding:"min=0"`
	TotalChunks  int    `form:"total_chunks" binding:"min=1"`
	OriginalName string `form:"original_name"`
}

type CommentCreateDTO struct {
	EpisodeID   uint64 `json:"episode_id"`
	ChannelID   uint64 `json:"channel_id"`
	ParentID    uint64 `json:"parent_id"`
	Type        string `json:"type"`
	Content     string `json:"content" binding:"required,min=1,max=5000"`
	Rating      int    `json:"rating" binding:"min=0,max=5"`
	IsDanmaku   bool   `json:"is_danmaku"`
	DanmakuTime int    `json:"danmaku_time"`
}

type CommentQuery struct {
	Page     int  `form:"page" binding:"min=1"`
	PageSize int  `form:"page_size" binding:"min=1,max=100"`
	Status   *int `form:"status"`
}

type SubscribeDTO struct {
	Email  string `json:"email" binding:"required,email,max=200"`
	Source string `json:"source" binding:"max=50"`
}
