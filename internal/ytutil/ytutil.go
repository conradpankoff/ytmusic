package ytutil

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/url"
	"regexp"
	"strings"

	"fknsrs.biz/p/ytmusic/internal/ctxhttpclient"
)

type IDType string

const (
	InvalidID  = IDType("invalid")
	ChannelID  = IDType("channel")
	PlaylistID = IDType("playlist")
	VideoID    = IDType("video")
)

type ID struct {
	Type  IDType
	Value string
}

func ExtractAndIdentifyIDs(text string, ignoreInvalid bool) ([]ID, error) {
	var ids []ID

	for _, urlOrID := range strings.Fields(text) {
		if idType, id, err := ExtractAndIdentifyID(urlOrID); err == nil {
			ids = append(ids, ID{idType, id})
		} else if !ignoreInvalid {
			return nil, fmt.Errorf("ytutil.ExtractAndIdentifyIDs: could not identify %q: %w", urlOrID, err)
		}
	}

	return ids, nil
}

func ExtractAndIdentifyID(urlOrID string) (IDType, string, error) {
	if channelID, err := ExtractChannelID(urlOrID); err == nil {
		return ChannelID, channelID, nil
	}

	if playlistID, err := ExtractPlaylistID(urlOrID); err == nil {
		return PlaylistID, playlistID, nil
	}

	if videoID, err := ExtractVideoID(urlOrID); err == nil {
		return VideoID, videoID, nil
	}

	return InvalidID, "", fmt.Errorf("ytutil.ExtractAndIdentifyID: could not extract a known ID type")
}

func ExtractChannelID(urlOrID string) (string, error) {
	if len(urlOrID) == 24 {
		return urlOrID, nil
	}

	if parsed, err := url.Parse(urlOrID); err == nil {
		if parsed.Path == "/channel" || strings.HasPrefix(parsed.Path, "/channel/") || parsed.Path == "/c" || strings.HasPrefix(parsed.Path, "/c/") {
			id := parsed.Query().Get("channel_id")

			if id == "" {
				parts := strings.Split(parsed.Path, "/")
				if len(parts) >= 3 {
					id = parts[2]
				}
			}

			if len(id) != 24 {
				return "", fmt.Errorf("ytutil.ExtractChannelID: invalid channel id; length should be 24")
			}

			return id, nil
		}
	}

	return "", fmt.Errorf("ytutil.ExtractChannelID: invalid url or id; could not find a known pattern")
}

func ExtractPlaylistID(urlOrID string) (string, error) {
	u, err := url.Parse(urlOrID)
	if err == nil && u.Scheme != "" && u.Host == "www.youtube.com" && u.Path == "/playlist" {
		return ExtractPlaylistID(u.Query().Get("list"))
	}

	playlistID := strings.TrimSpace(urlOrID)
	if len(playlistID) == 0 {
		return "", fmt.Errorf("ytutil.ExtractPlaylistID: empty input")
	}

	if strings.HasPrefix(playlistID, "PL") || strings.HasPrefix(playlistID, "UU") || strings.HasPrefix(playlistID, "FL") {
		return playlistID, nil
	}

	if len(playlistID) == 34 || len(playlistID) == 41 {
		return playlistID, nil
	}

	return "", fmt.Errorf("ytutil.ExtractPlaylistID: invalid url or id; could not find a known pattern")
}

func ExtractVideoID(urlOrID string) (string, error) {
	if len(urlOrID) == 11 {
		return urlOrID, nil
	}

	parsed, err := url.Parse(urlOrID)
	if err != nil {
		return "", err
	}

	if parsed.Host == "www.youtube.com" && parsed.Path == "/watch" {
		if id := parsed.Query().Get("v"); id != "" {
			if len(id) != 11 {
				return "", fmt.Errorf("invalid video id for v parameter in youtube.com url; length should be 11")
			}

			return id, nil
		}

		return "", fmt.Errorf("no v query parameter in youtube.com url")
	}

	if parsed.Host == "youtu.be" {
		if id := strings.TrimPrefix(parsed.Path, "/"); id != "" {
			if len(id) != 11 {
				return "", fmt.Errorf("invalid video id for youtu.be url; length should be 11")
			}

			return id, nil
		}

		return "", fmt.Errorf("no path content found in youtu.be url")
	}

	return "", fmt.Errorf("invalid url or id; could not find a known pattern")
}

func FindChannelID(ctx context.Context, urlOrID string) (string, error) {
	if idType, id, err := ExtractAndIdentifyID(urlOrID); err == nil {
		switch idType {
		case ChannelID:
			return id, nil
		case PlaylistID:
			channelID, err := getChannelIDFromURL(ctx, "https://www.youtube.com/playlist?list="+id)
			if err != nil {
				return "", fmt.Errorf("ytutil.FindChannelID: %w", err)
			}
			return channelID, nil
		case VideoID:
			channelID, err := getChannelIDFromURL(ctx, "https://www.youtube.com/watch?v="+id)
			if err != nil {
				return "", fmt.Errorf("ytutil.FindChannelID: %w", err)
			}
			return channelID, nil
		}
	}

	if strings.HasPrefix(urlOrID, "http:") || strings.HasPrefix(urlOrID, "https:") {
		return getChannelIDFromURL(ctx, urlOrID)
	}

	return "", fmt.Errorf("ytutil.FindChannelID: no strategy available to extract channel ID")
}

func getChannelIDFromURL(ctx context.Context, url string) (string, error) {
	res, err := ctxhttpclient.GetHTTPClient(ctx).Get(url)
	if err != nil {
		return "", fmt.Errorf("ytutil.getChannelIDFromURL: could not perform request: %w", err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("ytutil.getChannelIDFromURL: could not read response: %w", err)
	}

	re := regexp.MustCompile(`channelId=UC([-_a-zA-Z0-9]+)`)
	match := re.FindStringSubmatch(string(body))
	if len(match) > 1 {
		return "UC" + match[1], nil
	}

	return "", fmt.Errorf("ytutil.getChannelIDFromURL: could not find channel id in response")
}
