package utils

import (
	"os"
	"path/filepath"
	"testing"
)

// UploadChunk -> AudioService.MergeChunks -> RemoveFile temporary resource lifecycle
func TestRemoveFileRemovesTemporaryDirectory(t *testing.T) {
	dir := t.TempDir()
	chunkDir := filepath.Join(dir, "chunks")
	if err := os.Mkdir(chunkDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(chunkDir, "00000.part"), []byte("audio"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(chunkDir, "00000.part")); err != nil {
		t.Fatal(err)
	}
	if err := RemoveFile(chunkDir); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(chunkDir); !os.IsNotExist(err) {
		t.Fatalf("temporary directory was not removed: %v", err)
	}
}
