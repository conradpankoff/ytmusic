package ytdl

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
)

const (
	ProgramName = "yt-dlp"
)

func makeProcess(args []string) *exec.Cmd {
	return exec.Command(ProgramName, args...)
}

func runCommandAndGetJSON(args []string, output interface{}) error {
	stdout, err := makeProcess(args).Output()
	if err != nil {
		return err
	}

	if err := json.Unmarshal(stdout, output); err != nil {
		return fmt.Errorf("failed to unmarshal json: %w", err)
	}

	return nil
}

func runCommand(args []string) error {
	if _, err := makeProcess(args).Output(); err != nil {
		return err
	}

	return nil
}

type ProgressCallback func(progress int)

func DownloadVideo(ctx context.Context, id string, outputFile string) error {
	return DownloadVideoWithProgress(ctx, id, outputFile, nil)
}

func DownloadVideoWithProgress(ctx context.Context, id string, outputFile string, progressCallback ProgressCallback) error {
	cmd := exec.CommandContext(ctx, ProgramName,
		"-f", "bestvideo+bestaudio",
		"-S", "ext:mp4:m4a",
		"-o", outputFile,
		"--newline",
		"https://www.youtube.com/watch?v="+id,
	)

	if progressCallback == nil {
		// Use the simple version without progress tracking
		if _, err := cmd.Output(); err != nil {
			return fmt.Errorf("failed to download video: %w", err)
		}
		return nil
	}

	// Set up pipes for progress monitoring
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start download: %w", err)
	}

	// Progress pattern for yt-dlp: [download]  45.2% of  123.45MiB at    1.23MiB/s ETA 00:12
	progressPattern := regexp.MustCompile(`\[download\]\s+(\d+(?:\.\d+)?)%`)

	// Monitor stdout for progress
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			if matches := progressPattern.FindStringSubmatch(line); len(matches) > 1 {
				if percent, err := strconv.ParseFloat(matches[1], 32); err == nil {
					progressCallback(int(percent))
				}
			}
		}
	}()

	// Monitor stderr for errors
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			// Still look for progress in stderr as yt-dlp sometimes outputs there
			if matches := progressPattern.FindStringSubmatch(line); len(matches) > 1 {
				if percent, err := strconv.ParseFloat(matches[1], 32); err == nil {
					progressCallback(int(percent))
				}
			}
		}
	}()

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("failed to download video: %w", err)
	}

	return nil
}
