package service

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"podcast-platform/config"
	appErr "podcast-platform/pkg/errors"
	"podcast-platform/pkg/ffmpeg"
	"podcast-platform/pkg/logger"
	"podcast-platform/pkg/utils"
)

type AudioService struct {
	processor *ffmpeg.Processor
	storage   *StorageService
}

func NewAudioService(processor *ffmpeg.Processor, storage *StorageService) *AudioService {
	return &AudioService{processor: processor, storage: storage}
}

type UploadResult struct {
	FilePath     string
	FileURL      string
	FileSize     int64
	Duration     int
	SampleRate   int
	BitRate      int
	MimeType     string
	OriginalName string
	Title        string
	Author       string
}

type StorageService struct {
	storageType string
	localPath   string
	baseURL     string
}

func NewStorageService(cfg *config.StorageConfig, baseURL string) *StorageService {
	return &StorageService{
		storageType: cfg.Type,
		localPath:   cfg.Local.Path,
		baseURL:     baseURL,
	}
}

func (s *StorageService) SaveFile(src multipart.File, fileName string, subDir string) (string, string, int64, error) {
	dir := utils.GenerateDatePath(filepath.Join(s.localPath, subDir))
	if err := utils.EnsureDir(dir); err != nil {
		return "", "", 0, appErr.Wrap(err, 500, "create dir failed")
	}
	dstName := utils.GenerateFileName(fileName)
	dstPath := filepath.Join(dir, dstName)
	dst, err := os.Create(dstPath)
	if err != nil {
		return "", "", 0, appErr.Wrap(err, 500, "create file failed")
	}
	defer dst.Close()
	size, err := io.Copy(dst, src)
	if err != nil {
		return "", "", 0, appErr.Wrap(err, 500, "copy file failed")
	}
	relPath, _ := filepath.Rel(s.localPath, dstPath)
	urlPath := "/storage/" + filepath.ToSlash(relPath)
	return dstPath, urlPath, size, nil
}

func (s *StorageService) ServePath() string {
	return s.localPath
}

func (s *AudioService) UploadAudio(fileHeader *multipart.FileHeader) (*UploadResult, error) {
	if fileHeader == nil {
		return nil, appErr.ErrInvalidParams
	}
	if !utils.IsAudioFile(fileHeader.Filename) {
		return nil, appErr.ErrInvalidAudio
	}
	src, err := fileHeader.Open()
	if err != nil {
		return nil, appErr.Wrap(err, 400, "open upload file")
	}
	defer src.Close()

	storagePath := config.AppConfig.Storage.Local.Path
	storageSvc := NewStorageService(&config.AppConfig.Storage, "")
	_ = storageSvc

	dir := utils.GenerateDatePath(filepath.Join(storagePath, "audio"))
	if err := utils.EnsureDir(dir); err != nil {
		return nil, appErr.Wrap(err, 500, "create dir")
	}
	dstName := utils.GenerateFileName(fileHeader.Filename)
	dstPath := filepath.Join(dir, dstName)
	dst, err := os.Create(dstPath)
	if err != nil {
		return nil, appErr.Wrap(err, 500, "create file")
	}
	defer dst.Close()
	size, err := io.Copy(dst, src)
	if err != nil {
		return nil, appErr.Wrap(err, 500, "copy file")
	}
	relPath, _ := filepath.Rel(storagePath, dstPath)
	urlPath := "/storage/" + filepath.ToSlash(relPath)

	result := &UploadResult{
		FilePath:     dstPath,
		FileURL:      urlPath,
		FileSize:     size,
		OriginalName: fileHeader.Filename,
		MimeType:     s.processor.GetMimeType(utils.GetExt(fileHeader.Filename)),
	}

	if s.processor.Available() {
		info, err := s.processor.GetAudioInfo(dstPath)
		if err == nil && info != nil {
			result.Duration = info.Duration
			result.SampleRate = info.SampleRate
			result.BitRate = info.BitRate
			result.Title = info.Title
			result.Author = info.Author
			if info.FileSize > 0 {
				result.FileSize = info.FileSize
			}
			if info.Codec != "" {
				logger.Infof("audio codec: %s, duration: %ds", info.Codec, info.Duration)
			}
		} else {
			logger.Warnf("ffprobe failed: %v", err)
		}
	}
	return result, nil
}

func (s *AudioService) UploadChunked(fileHeader *multipart.FileHeader, chunkIndex, totalChunks int, uploadID string) (bool, string, error) {
	if fileHeader == nil {
		return false, "", appErr.ErrInvalidParams
	}
	src, err := fileHeader.Open()
	if err != nil {
		return false, "", appErr.Wrap(err, 400, "open chunk")
	}
	defer src.Close()

	tempDir := filepath.Join(config.AppConfig.Storage.Local.Path, "chunks", uploadID)
	if err := utils.EnsureDir(tempDir); err != nil {
		return false, "", appErr.Wrap(err, 500, "chunk dir")
	}
	chunkPath := filepath.Join(tempDir, fmt.Sprintf("%05d.part", chunkIndex))
	dst, err := os.Create(chunkPath)
	if err != nil {
		return false, "", appErr.Wrap(err, 500, "create chunk")
	}
	defer dst.Close()
	_, err = io.Copy(dst, src)
	if err != nil {
		return false, "", appErr.Wrap(err, 500, "write chunk")
	}

	if chunkIndex+1 >= totalChunks {
		return true, tempDir, nil
	}
	return false, "", nil
}

func (s *AudioService) MergeChunks(tempDir, originalName string) (*UploadResult, error) {
	audioDir := utils.GenerateDatePath(filepath.Join(config.AppConfig.Storage.Local.Path, "audio"))
	if err := utils.EnsureDir(audioDir); err != nil {
		return nil, appErr.Wrap(err, 500, "audio dir")
	}
	dstName := utils.GenerateFileName(originalName)
	dstPath := filepath.Join(audioDir, dstName)
	dst, err := os.Create(dstPath)
	if err != nil {
		return nil, appErr.Wrap(err, 500, "create merged")
	}
	defer dst.Close()
	defer utils.RemoveFile(tempDir)

	files, _ := filepath.Glob(filepath.Join(tempDir, "*.part"))
	for i := 0; i < len(files); i++ {
		p := filepath.Join(tempDir, fmt.Sprintf("%05d.part", i))
		if !utils.FileExists(p) {
			continue
		}
		part, err := os.Open(p)
		if err != nil {
			continue
		}
		_, _ = io.Copy(dst, part)
		part.Close()
		_ = os.Remove(p)
	}

	size, _ := utils.FileSize(dstPath)
	relPath, _ := filepath.Rel(config.AppConfig.Storage.Local.Path, dstPath)
	urlPath := "/storage/" + filepath.ToSlash(relPath)
	result := &UploadResult{
		FilePath:     dstPath,
		FileURL:      urlPath,
		FileSize:     size,
		OriginalName: originalName,
		MimeType:     s.processor.GetMimeType(utils.GetExt(originalName)),
	}
	if s.processor.Available() {
		if info, err := s.processor.GetAudioInfo(dstPath); err == nil {
			result.Duration = info.Duration
			result.SampleRate = info.SampleRate
			result.BitRate = info.BitRate
			result.Title = info.Title
			result.Author = info.Author
		}
	}
	return result, nil
}
