package ffmpeg

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	appErr "podcast-platform/pkg/errors"
)

type AudioInfo struct {
	Duration   int    `json:"duration"`
	SampleRate int    `json:"sample_rate"`
	BitRate    int    `json:"bit_rate"`
	Channels   int    `json:"channels"`
	Codec      string `json:"codec"`
	Format     string `json:"format"`
	FileSize   int64  `json:"file_size"`
	Title      string `json:"title"`
	Author     string `json:"author"`
	Album      string `json:"album"`
}

type Processor struct {
	FFmpegPath  string
	FFprobePath string
}

func NewProcessor() *Processor {
	ffmpeg, _ := exec.LookPath("ffmpeg")
	ffprobe, _ := exec.LookPath("ffprobe")
	if ffmpeg == "" {
		ffmpeg = "ffmpeg"
	}
	if ffprobe == "" {
		ffprobe = "ffprobe"
	}
	return &Processor{
		FFmpegPath:  ffmpeg,
		FFprobePath: ffprobe,
	}
}

func (p *Processor) GetAudioInfo(filePath string) (*AudioInfo, error) {
	cmd := exec.Command(p.FFprobePath,
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		filePath,
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, appErr.Wrap(err, 400, "failed to probe audio file")
	}

	var probe struct {
		Streams []struct {
			CodecName  string `json:"codec_name"`
			SampleRate string `json:"sample_rate"`
			Channels   int    `json:"channels"`
			BitRate    string `json:"bit_rate"`
			Duration   string `json:"duration"`
		} `json:"streams"`
		Format struct {
			Duration string `json:"duration"`
			Size     string `json:"size"`
			BitRate  string `json:"bit_rate"`
			Tags     struct {
				Title  string `json:"title"`
				Artist string `json:"artist"`
				Album  string `json:"album"`
			} `json:"tags"`
		} `json:"format"`
	}

	if err := json.Unmarshal(output, &probe); err != nil {
		return nil, appErr.Wrap(err, 400, "invalid audio metadata")
	}

	info := &AudioInfo{
		Title:  probe.Format.Tags.Title,
		Author: probe.Format.Tags.Artist,
		Album:  probe.Format.Tags.Album,
	}

	if dur, err := strconv.ParseFloat(probe.Format.Duration, 64); err == nil {
		info.Duration = int(dur)
	}
	if size, err := strconv.ParseInt(probe.Format.Size, 10, 64); err == nil {
		info.FileSize = size
	}
	if br, err := strconv.Atoi(probe.Format.BitRate); err == nil {
		info.BitRate = br
	}

	if len(probe.Streams) > 0 {
		s := probe.Streams[0]
		info.Codec = s.CodecName
		info.Channels = s.Channels
		if sr, err := strconv.Atoi(s.SampleRate); err == nil {
			info.SampleRate = sr
		}
		if s.BitRate != "" {
			if br, err := strconv.Atoi(s.BitRate); err == nil {
				info.BitRate = br
			}
		}
		if info.Duration == 0 {
			if dur, err := strconv.ParseFloat(s.Duration, 64); err == nil {
				info.Duration = int(dur)
			}
		}
	}

	return info, nil
}

func (p *Processor) ConvertToMP3(input, output string, bitRate int) error {
	if bitRate <= 0 {
		bitRate = 128
	}
	args := []string{
		"-i", input,
		"-codec:a", "libmp3lame",
		"-b:a", fmt.Sprintf("%dk", bitRate),
		"-y",
		output,
	}
	cmd := exec.Command(p.FFmpegPath, args...)
	if err := cmd.Run(); err != nil {
		return appErr.Wrap(err, 400, "audio conversion failed")
	}
	return nil
}

func (p *Processor) GetMimeType(ext string) string {
	switch strings.ToLower(ext) {
	case "mp3":
		return "audio/mpeg"
	case "m4a", "aac":
		return "audio/mp4"
	case "wav":
		return "audio/wav"
	case "flac":
		return "audio/flac"
	case "ogg", "opus":
		return "audio/ogg"
	case "wma":
		return "audio/x-ms-wma"
	default:
		return "audio/mpeg"
	}
}

func (p *Processor) Available() bool {
	_, err1 := exec.LookPath(p.FFmpegPath)
	_, err2 := exec.LookPath(p.FFprobePath)
	return err1 == nil && err2 == nil
}
