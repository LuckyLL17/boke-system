package service

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"podcast-platform/config"
	"podcast-platform/pkg/ffmpeg"
	"podcast-platform/pkg/utils"
)

// setTempStorage points the global storage path at a temp dir for the test and
// returns it. It panics on failure so the test fails fast.
func setTempStorage(t *testing.T) string {
	t.Helper()
	if config.AppConfig == nil {
		config.AppConfig = &config.Config{}
	}
	dir := t.TempDir()
	config.AppConfig.Storage.Local.Path = dir
	return dir
}

func newTestAudioService() *AudioService {
	return NewAudioService(ffmpeg.NewProcessor(), nil)
}

// makeChunkHeader builds a multipart.FileHeader in memory whose body is body.
func makeChunkHeader(t *testing.T, body []byte) *multipart.FileHeader {
	t.Helper()
	// Create via a temp file and wrap using multipartFileInfo through a writer.
	// Simpler: build a multipart body and parse it.
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, err := mw.CreateFormFile("file", "chunk.bin")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := fw.Write(body); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	if err := mw.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	if err := req.ParseMultipartForm(32 << 20); err != nil {
		t.Fatalf("parse multipart form: %v", err)
	}
	fh := req.MultipartForm.File["file"][0]
	return fh
}

// TestMergeChunks_OutOfOrder verifies that chunks arriving out of order are
// assembled in the correct order and that the merged output byte-for-byte
// equals the original (including the tail). The pre-fix code iterated the
// merge loop from 0 with a mismatched filename format, dropping the first
// chunk's data and never copying the last — truncating the tail.
func TestMergeChunks_OutOfOrder(t *testing.T) {
	setTempStorage(t)
	svc := newTestAudioService()

	// Build a deterministic payload and split into 5 unequal chunks.
	total := 5
	original := bytes.Repeat([]byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"), 64)
	chunkSize := len(original) / total
	var chunks [][]byte
	for i := 0; i < total; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if i == total-1 {
			end = len(original) // last chunk gets the remainder (the "tail")
		}
		chunks = append(chunks, original[start:end])
	}

	uploadID := "order-test"

	// Upload in reverse order: last chunk first, earliest chunk last. This is
	// exactly the scenario the bug report describes (out-of-order arrival where
	// the last request triggers merge).
	order := []int{4, 2, 0, 3, 1}
	var lastTempDir string
	var doneSeen bool
	for _, idx := range order {
		fh := makeChunkHeader(t, chunks[idx])
		done, tempDir, err := svc.UploadChunked(fh, idx, total, uploadID)
		if err != nil {
			t.Fatalf("UploadChunked(idx=%d): %v", idx, err)
		}
		lastTempDir = tempDir
		if done {
			doneSeen = true
		}
	}
	if !doneSeen {
		t.Fatalf("expected merge to be triggered after all chunks arrived, but done was never true")
	}

	result, err := svc.MergeChunks(lastTempDir, "recording.mp3", total, uploadID)
	if err != nil {
		t.Fatalf("MergeChunks: %v", err)
	}

	merged, err := os.ReadFile(result.FilePath)
	if err != nil {
		t.Fatalf("read merged file: %v", err)
	}
	if !bytes.Equal(merged, original) {
		t.Fatalf("merged content does not match original (got %d bytes, want %d); tail corruption",
			len(merged), len(original))
	}

	// Verify the temp chunk directory was cleaned up on success.
	if utils.FileExists(filepath.Join(config.AppConfig.Storage.Local.Path, "chunks", uploadID)) {
		t.Fatalf("chunk temp dir should be removed after successful merge")
	}
}

// TestMergeChunks_IncompleteAborts verifies that when chunks are missing the
// merge refuses to report success and cleans up the partial state.
func TestMergeChunks_IncompleteAborts(t *testing.T) {
	setTempStorage(t)
	svc := newTestAudioService()

	total := 4
	uploadID := "incomplete-test"

	// Only upload chunks 0 and 2; leave 1 and 3 missing.
	for _, idx := range []int{0, 2} {
		fh := makeChunkHeader(t, bytes.Repeat([]byte{byte(idx)}, 16))
		if _, _, err := svc.UploadChunked(fh, idx, total, uploadID); err != nil {
			t.Fatalf("UploadChunked(idx=%d): %v", idx, err)
		}
	}

	tempDir := filepath.Join(config.AppConfig.Storage.Local.Path, "chunks", uploadID)
	_, err := svc.MergeChunks(tempDir, "recording.mp3", total, uploadID)
	if err == nil {
		t.Fatalf("expected MergeChunks to fail when chunks are missing, but it succeeded")
	}

	// Temp dir should be cleaned up even on the incomplete-chunks failure path.
	if utils.FileExists(tempDir) {
		t.Fatalf("chunk temp dir should be removed on incomplete-chunks failure")
	}
}

// TestMergeChunks_ConcurrentNoDoubleMerge verifies that two concurrent merge
// attempts on the same complete upload do not both report success (no
// duplicate / partial files) and that the surviving output is correct. The
// per-upload lock makes exactly one merge win.
func TestMergeChunks_ConcurrentNoDoubleMerge(t *testing.T) {
	setTempStorage(t)
	svc := newTestAudioService()

	total := 4
	uploadID := "concurrent-test"
	original := bytes.Repeat([]byte("concurrent-merge-payload-"), 37)
	cs := len(original) / total
	chunks := [][]byte{
		original[0:cs],
		original[cs : 2*cs],
		original[2*cs : 3*cs],
		original[3*cs:],
	}

	tempDir := filepath.Join(config.AppConfig.Storage.Local.Path, "chunks", uploadID)
	for i := 0; i < total; i++ {
		fh := makeChunkHeader(t, chunks[i])
		if _, _, err := svc.UploadChunked(fh, i, total, uploadID); err != nil {
			t.Fatalf("UploadChunked(idx=%d): %v", i, err)
		}
	}

	type res struct {
		path string
		err  error
	}
	resCh := make(chan res, 2)
	for i := 0; i < 2; i++ {
		go func() {
			r, err := svc.MergeChunks(tempDir, "concurrent.mp3", total, uploadID)
			resCh <- res{path: pathOf(r), err: err}
		}()
	}

	var successes []string
	for i := 0; i < 2; i++ {
		r := <-resCh
		if r.err == nil {
			successes = append(successes, r.path)
		}
	}

	if len(successes) != 1 {
		t.Fatalf("expected exactly one successful merge, got %d (paths=%v)", len(successes), successes)
	}

	merged, err := os.ReadFile(successes[0])
	if err != nil {
		t.Fatalf("read merged file: %v", err)
	}
	if !bytes.Equal(merged, original) {
		t.Fatalf("merged content does not match original (got %d bytes, want %d)",
			len(merged), len(original))
	}

	// Only the single merged file should exist under audio/ (no duplicate).
	audioDir := filepath.Join(config.AppConfig.Storage.Local.Path, "audio")
	count := 0
	err = filepath.WalkDir(audioDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !d.IsDir() {
			count++
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk audio dir: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected exactly 1 merged file, found %d", count)
	}
}

func pathOf(r *UploadResult) string {
	if r == nil {
		return ""
	}
	return r.FilePath
}
