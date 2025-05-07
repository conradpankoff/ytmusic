package ffmpeg

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

func MakeThumbnail(ctx context.Context, videoFile, imageFile string) (string, error) {
	cmd := exec.CommandContext(
		ctx, "ffmpeg",
		"-y",
		"-loglevel", "warning",
		"-i", videoFile,
		"-ss", "00:00:01",
		"-vframes", "1",
		imageFile,
	)

	var buf bytes.Buffer

	cmd.Stdin = nil
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	if err := cmd.Run(); err != nil {
		return buf.String(), fmt.Errorf("ffmpeg.MakeThumbnail: %w", err)
	}

	return buf.String(), nil
}

func Transcode(ctx context.Context, inputFile, size, outputFile string) (string, error) {
	cmd := exec.CommandContext(
		ctx, "ffmpeg",
		"-y",
		"-loglevel", "warning",
		"-i", inputFile,
		"-vf", "scale="+size,
		"-c:v", "libx264",
		"-crf", "23",
		"-tune", "fastdecode",
		"-preset", "veryslow",
		"-c:a", "libmp3lame",
		"-q:a", "2",
		outputFile,
	)

	var buf bytes.Buffer

	cmd.Stdin = nil
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	if err := cmd.Run(); err != nil {
		return buf.String(), fmt.Errorf("ffmpeg.Transcode: %w", err)
	}

	return buf.String(), nil
}

func ExtractAudio(ctx context.Context, videoFile, audioFile string) (string, error) {
	cmd := exec.CommandContext(
		ctx, "ffmpeg",
		"-y",
		"-loglevel", "warning",
		"-i", videoFile,
		"-vn",
		"-c:a", "libmp3lame",
		"-q:a", "2",
		audioFile,
	)

	var buf bytes.Buffer

	cmd.Stdin = nil
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	if err := cmd.Run(); err != nil {
		return buf.String(), fmt.Errorf("ffmpeg.ExtractAudio: %w", err)
	}

	return buf.String(), nil
}
