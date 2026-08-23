package service

import (
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"podcast-platform/config"
	appErr "podcast-platform/pkg/errors"
	"podcast-platform/pkg/ffmpeg"
	"podcast-platform/pkg/logger"
	"podcast-platform/pkg/utils"
)

type AudioService struct {
	processor *ffmpeg.Processor
	storage   *StorageService

	uploadsMu sync.Mutex
	locks     map[string]*sync.Mutex
}

func NewAudioService(processor *ffmpeg.Processor, storage *StorageService) *AudioService {
	return &AudioService{processor: processor, storage: storage, locks: make(map[string]*sync.Mutex)}
}

// uploadLock returns a per-upload-ID mutex so concurrent chunk requests for the
// same file (which may arrive out of order) are serialized, preventing a merge
// from racing with a chunk still being written or a merge running twice.
func (s *AudioService) uploadLock(uploadID string) *sync.Mutex {
	s.uploadsMu.Lock()
	defer s.uploadsMu.Unlock()
	mu, ok := s.locks[uploadID]
	if !ok {
		mu = &sync.Mutex{}
		s.locks[uploadID] = mu
	}
	return mu
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
		return false, "", appErr.Wrap(err, http.StatusBadRequest, "open chunk")
	}
	defer src.Close()

	tempDir := filepath.Join(config.AppConfig.Storage.Local.Path, "chunks", uploadID)
	if err := utils.EnsureDir(tempDir); err != nil {
		return false, "", appErr.Wrap(err, http.StatusInternalServerError, "chunk dir")
	}

	// Serialize all chunk writes for this upload: chunks may arrive out of
	// order or concurrently, and the merge decision below must see a stable
	// on-disk view of which parts already exist.
	mu := s.uploadLock(uploadID)
	mu.Lock()
	defer mu.Unlock()

	// Write the chunk body to a temp file in the same directory, then rename
	// atomically so a partially written .part is never observed by the merge.
	chunkPath := utils.ChunkPartPath(tempDir, chunkIndex)
	tmpPath := chunkPath + ".tmp"
	dst, err := os.Create(tmpPath)
	if err != nil {
		return false, "", appErr.Wrap(err, http.StatusInternalServerError, "create chunk")
	}
	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		_ = os.Remove(tmpPath)
		return false, "", appErr.Wrap(err, http.StatusInternalServerError, "write chunk")
	}
	if err := dst.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return false, "", appErr.Wrap(err, http.StatusInternalServerError, "close chunk")
	}
	if err := os.Rename(tmpPath, chunkPath); err != nil {
		_ = os.Remove(tmpPath)
		return false, "", appErr.Wrap(err, http.StatusInternalServerError, "persist chunk")
	}

	// Report "done" only when every expected chunk now exists on disk — not
	// merely when the highest-indexed chunk has arrived. This is what makes
	// out-of-order arrivals safe: a late earlier chunk that lands after the
	// last-indexed one will still trigger (or re-trigger) the merge once the
	// set becomes complete. The handler tolerates a non-final chunk also
	// returning done=true as long as all parts are present.
	if !s.allChunksPresent(tempDir, totalChunks) {
		return false, tempDir, nil
	}
	return true, tempDir, nil
}

// allChunksPresent reports whether every chunk 0..totalChunks-1 exists and is
// non-empty on disk. A zero-size part is treated as missing so a truncated
// write can never satisfy the completeness check.
func (s *AudioService) allChunksPresent(tempDir string, totalChunks int) bool {
	for i := 0; i < totalChunks; i++ {
		p := utils.ChunkPartPath(tempDir, i)
		info, err := os.Stat(p)
		if err != nil || info.Size() == 0 {
			return false
		}
	}
	return true
}

func (s *AudioService) MergeChunks(tempDir, originalName string, totalChunks int, uploadID string) (*UploadResult, error) {
	// Hold the per-upload lock across the whole merge so a second concurrent
	// "done" request (possible when chunks arrive out of order and the set
	// becomes complete on more than one of them) cannot run a duplicate merge.
	mu := s.uploadLock(uploadID)
	mu.Lock()
	defer mu.Unlock()

	// Integrity gate: refuse to merge unless every expected chunk is present and
	// non-empty. This is the guarantee that the assembled file contains the full
	// original content before success is ever reported. If a prior merge already
	// consumed the temp dir, it is absent here and we surface a clean error
	// rather than producing a partial file.
	if totalChunks <= 0 || !s.allChunksPresent(tempDir, totalChunks) {
		_ = os.RemoveAll(tempDir)
		return nil, appErr.Wrap(os.ErrNotExist, http.StatusBadRequest, "incomplete chunks, upload aborted")
	}

	audioDir := utils.GenerateDatePath(filepath.Join(config.AppConfig.Storage.Local.Path, "audio"))
	if err := utils.EnsureDir(audioDir); err != nil {
		return nil, appErr.Wrap(err, http.StatusInternalServerError, "audio dir")
	}
	dstName := utils.GenerateFileName(originalName)
	dstPath := filepath.Join(audioDir, dstName)
	dst, err := os.Create(dstPath)
	if err != nil {
		_ = os.RemoveAll(tempDir)
		return nil, appErr.Wrap(err, http.StatusInternalServerError, "create merged")
	}

	// Assemble in chunk-index order (00001.part, 00002.part, ...). The previous
	// loop iterated from 0 and used a format string that did not match the
	// ChunkPartPath naming (index+1), so the first part was silently skipped and
	// the last part was never copied — truncating the tail of the audio.
	merged := int64(0)
	var copyErr error
	for i := 0; i < totalChunks; i++ {
		p := utils.ChunkPartPath(tempDir, i)
		if err := copyPart(dst, p); err != nil {
			copyErr = err
			break
		}
		merged++
	}
	dst.Close()

	if copyErr != nil {
		// Failure: remove the partial merged file so it is never served, and
		// clean up the chunk directory too (retain the failure-cleanup behavior).
		_ = os.Remove(dstPath)
		_ = os.RemoveAll(tempDir)
		return nil, appErr.Wrap(copyErr, http.StatusInternalServerError, "merge chunk "+strconv.FormatInt(merged, 10)+" failed")
	}

	// Success: remove the per-chunk parts and the temp directory.
	_ = os.RemoveAll(tempDir)

	size, _ := utils.FileSize(dstPath)
	if size == 0 {
		_ = os.Remove(dstPath)
		return nil, appErr.Wrap(os.ErrNotExist, http.StatusInternalServerError, "merged file is empty")
	}
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

// copyPart appends the contents of the part file at partPath to dst. It opens
// the part fresh for each chunk so memory usage stays bounded regardless of
// upload size.
func copyPart(dst io.Writer, partPath string) error {
	f, err := os.Open(partPath)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := io.Copy(dst, f); err != nil {
		return err
	}
	return nil
}
