package ffmpeg

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
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

type ProgressCallback func(progress int)

func Transcode(ctx context.Context, inputFile, size, outputFile string) (string, error) {
	return TranscodeWithProgress(ctx, inputFile, size, outputFile, nil)
}

func TranscodeWithProgress(ctx context.Context, inputFile, size, outputFile string, progressCallback ProgressCallback) (string, error) {
	// First, get the duration of the input file for progress calculation
	duration, err := getVideoDuration(ctx, inputFile)
	if err != nil {
		return "", fmt.Errorf("failed to get video duration: %w", err)
	}

	cmd := exec.CommandContext(
		ctx, "ffmpeg",
		"-y",
		"-progress", "pipe:1",
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

	if progressCallback == nil {
		// Use the simple version without progress tracking
		var buf bytes.Buffer
		cmd.Stdin = nil
		cmd.Stdout = &buf
		cmd.Stderr = &buf

		if err := cmd.Run(); err != nil {
			return buf.String(), fmt.Errorf("ffmpeg.Transcode: %w", err)
		}
		return buf.String(), nil
	}

	// Set up pipes for progress monitoring
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	var output bytes.Buffer

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start transcode: %w", err)
	}

	// Monitor stdout for progress
	go func() {
		scanner := bufio.NewScanner(stdout)
		timePattern := regexp.MustCompile(`out_time_ms=(\d+)`)
		
		for scanner.Scan() {
			line := scanner.Text()
			if matches := timePattern.FindStringSubmatch(line); len(matches) > 1 {
				if timeMicros, err := strconv.ParseInt(matches[1], 10, 64); err == nil {
					currentTime := time.Duration(timeMicros) * time.Microsecond
					if duration > 0 {
						progress := int((currentTime.Seconds() / duration.Seconds()) * 100)
						if progress > 100 {
							progress = 100
						}
						if progress >= 0 {
							progressCallback(progress)
						}
					}
				}
			}
		}
	}()

	// Monitor stderr for output and errors
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			output.WriteString(line + "\n")
		}
	}()

	if err := cmd.Wait(); err != nil {
		return output.String(), fmt.Errorf("ffmpeg.Transcode: %w", err)
	}

	return output.String(), nil
}

// getVideoDuration extracts the duration of a video file using ffprobe
func getVideoDuration(ctx context.Context, inputFile string) (time.Duration, error) {
	cmd := exec.CommandContext(
		ctx, "ffprobe",
		"-v", "quiet",
		"-show_entries", "format=duration",
		"-of", "csv=p=0",
		inputFile,
	)

	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to get duration: %w", err)
	}

	durationStr := strings.TrimSpace(string(output))
	if durationSeconds, err := strconv.ParseFloat(durationStr, 64); err == nil {
		return time.Duration(durationSeconds * float64(time.Second)), nil
	}

	return 0, fmt.Errorf("failed to parse duration: %s", durationStr)
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
