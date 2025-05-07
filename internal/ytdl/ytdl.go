package ytdl

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
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

func DownloadVideo(ctx context.Context, id string, outputFile string) error {
	if err := runCommand([]string{
		"-f", "bestvideo+bestaudio",
		"-S", "ext:mp4:m4a",
		"-o", outputFile,
		"https://www.youtube.com/watch?v=" + id,
	}); err != nil {
		return fmt.Errorf("failed to download video: %w", err)
	}

	return nil
}
