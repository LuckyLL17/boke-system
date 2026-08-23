package ffmpeg

import (
	"testing"
)

// UploadAudio -> UploadChunked -> MergeChunks -> GetAudioInfo.
func TestAudioProbePreservesCommandFailure(t *testing.T) {
	processor := &Processor{FFprobePath: "false", FFmpegPath: "false"}
	_, err := processor.GetAudioInfo(t.TempDir() + "/missing.mp3")
	if err == nil {
		t.Fatalf("audio probe failure was hidden")
	}
}
