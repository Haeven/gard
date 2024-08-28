// internalencoder/encode.go
package encoder

import (
	"os/exec"
)

// EncodeVideo encodes a video file using the VP9 codec
func EncodeVideo(inputPath, outputPath string) error {
	cmd := exec.Command("ffmpeg", "-i", inputPath, "-c:v", "libvpx-vp9", outputPath)
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
