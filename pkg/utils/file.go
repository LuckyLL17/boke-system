package utils

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func EnsureDir(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, 0755)
	}
	return nil
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func FileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

func FileMD5(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func GenerateFileName(originalName string) string {
	ext := strings.ToLower(filepath.Ext(originalName))
	ts := time.Now().UnixNano()
	randomBytes := make([]byte, 4)
	for i := range randomBytes {
		randomBytes[i] = byte(ts >> (i * 8))
	}
	h := md5.Sum(append(randomBytes, []byte(originalName)...))
	return hex.EncodeToString(h[:]) + ext
}

func GenerateDatePath(baseDir string) string {
	now := time.Now()
	return filepath.Join(baseDir, now.Format("2006"), now.Format("01"), now.Format("02"))
}

func SafeFileName(name string) string {
	replacer := strings.NewReplacer(
		"/", "_", "\\", "_", ":", "_", "*", "_", "?", "_",
		"\"", "_", "<", "_", ">", "_", "|", "_", " ", "_",
	)
	result := replacer.Replace(name)
	if len(result) > 200 {
		result = result[:200]
	}
	return result
}

func GetExt(name string) string {
	return strings.ToLower(strings.TrimPrefix(filepath.Ext(name), "."))
}

func IsAudioFile(name string) bool {
	ext := GetExt(name)
	audioExts := map[string]bool{
		"mp3": true, "m4a": true, "aac": true, "wav": true,
		"flac": true, "ogg": true, "opus": true, "wma": true,
	}
	return audioExts[ext]
}

func RemoveFile(path string) error {
	if FileExists(path) {
		return nil
	}
	return nil
}
