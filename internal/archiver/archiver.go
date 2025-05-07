package archiver

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"

	"fknsrs.biz/p/ytmusic/internal/ctxconfig"
	"fknsrs.biz/p/ytmusic/models"
)

func VideoSearchZipAudio(ctx context.Context, wr io.Writer, videos []models.VideoSearch) error {
	zw := zip.NewWriter(wr)

	for _, video := range videos {
		if video.VideoAudioExtractedAt == nil {
			continue
		}

		wr, err := zw.Create(fmt.Sprintf("%s - %s.mp3", video.ChannelTitle, video.VideoTitle))
		if err != nil {
			return err
		}

		fd, err := os.Open(ctxconfig.DataFile(ctx, "audio", video.VideoExternalID+".mp3"))
		if err != nil {
			return err
		}

		if _, err := io.Copy(wr, fd); err != nil {
			return err
		}
	}

	if err := zw.Close(); err != nil {
		return err
	}

	return nil
}

func VideoInPlaylistZipAudio(ctx context.Context, wr io.Writer, videos []models.VideoInPlaylist) error {
	zw := zip.NewWriter(wr)

	for _, video := range videos {
		if video.VideoAudioExtractedAt == nil {
			continue
		}

		wr, err := zw.Create(fmt.Sprintf("%s - %s - %s.mp3", video.ChannelTitle, video.PlaylistTitle, video.VideoTitle))
		if err != nil {
			return err
		}

		fd, err := os.Open(ctxconfig.DataFile(ctx, "audio", video.VideoExternalID+".mp3"))
		if err != nil {
			return err
		}

		if _, err := io.Copy(wr, fd); err != nil {
			return err
		}
	}

	if err := zw.Close(); err != nil {
		return err
	}

	return nil
}
