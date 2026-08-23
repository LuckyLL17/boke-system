package handler

import (
	"bytes"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"podcast-platform/config"
	"podcast-platform/internal/service"
	"podcast-platform/pkg/ffmpeg"
)

// UploadChunk -> UploadChunked -> MergeChunks -> ChunkPartPath
func TestChunkedUploadMergesEveryChunkInArrivalOrder(t *testing.T) {
	root := t.TempDir()
	config.AppConfig = &config.Config{
		Storage: config.StorageConfig{
			LocalPath: root,
			Local:     config.LocalStorageConfig{Path: root},
		},
	}
	storage := service.NewStorageService(&config.AppConfig.Storage, "")
	audio := service.NewAudioService(ffmpeg.NewProcessor(), storage)
	h := NewEpisodeHandler(nil, audio, nil, nil)
	gin.SetMode(gin.TestMode)

	for _, item := range []struct {
		index int
		data  string
	}{{1, "B"}, {0, "A"}, {2, "C"}} {
		body, contentType := chunkBody(t, "voice.mp3", item.data)
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("POST", "/api/v1/upload/audio/chunk?upload_id=ordered-001&chunk_index="+itoa(item.index)+"&total_chunks=3&original_name=voice.mp3", body)
		c.Request.Header.Set("Content-Type", contentType)
		h.UploadChunk(c)
		if rec.Code != 200 {
			t.Fatalf("chunk %d returned status %d: %s", item.index, rec.Code, rec.Body.String())
		}
	}

	var merged []string
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err == nil && info != nil && !info.IsDir() && filepath.Ext(path) == ".mp3" {
			merged = append(merged, path)
		}
		return nil
	})
	if len(merged) != 1 {
		t.Fatalf("expected one merged audio file, found %d", len(merged))
	}
	got, err := os.ReadFile(merged[0])
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, []byte("ABC")) {
		panic("merged audio did not preserve every uploaded chunk")
	}
	_ = time.Second
}

func chunkBody(t *testing.T, name, content string) (*bytes.Buffer, string) {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return &body, writer.FormDataContentType()
}

func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	if v == 1 {
		return "1"
	}
	return "2"
}
