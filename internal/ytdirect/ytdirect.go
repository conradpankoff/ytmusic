package ytdirect

import (
  "context"
  "fmt"
  "io"
  "net/http"
  "strings"

  "github.com/Jeffail/gabs/v2"
  "github.com/PuerkitoBio/goquery"
  "golang.org/x/net/html"

  "fknsrs.biz/p/ytmusic/internal/ctxhttpclient"
)

func getData(ctx context.Context, url string) (io.ReadCloser, error) {
  req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
  if err != nil {
    return nil, fmt.Errorf("ytdirect.getData: %w", err)
  }

  res, err := ctxhttpclient.GetHTTPClient(ctx).Do(req)
  if err != nil {
    return nil, fmt.Errorf("ytdirect.getData: %w", err)
  }

  if res.StatusCode != http.StatusOK {
    return nil, fmt.Errorf("ytdirect.getData: status code: %d", res.StatusCode)
  }

  return res.Body, nil
}

func getDocument(ctx context.Context, url string) (*goquery.Document, error) {
  req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
  if err != nil {
    return nil, fmt.Errorf("ytdirect.getDocument: %w", err)
  }

  res, err := ctxhttpclient.GetHTTPClient(ctx).Do(req)
  if err != nil {
    return nil, fmt.Errorf("ytdirect.getDocument: %w", err)
  }

  doc, err := goquery.NewDocumentFromResponse(res)
  if err != nil {
    return nil, fmt.Errorf("ytdirect.getDocument: %w", err)
  }

  return doc, nil
}

type Channel struct {
  ID      string
  Title   string
  Shelves []ChannelShelf
}

type ChannelShelf struct {
  Title     string
  Playlists []ChannelPlaylist
}

type ChannelPlaylist struct {
  ID            string
  ChannelID     string
  Title         string
  PublishedTime string
  VideoCount    string
}

func GetChannel(ctx context.Context, id string) (*Channel, error) {
  doc, err := getDocument(ctx, "https://www.youtube.com/channel/"+id)
  if err != nil {
    return nil, fmt.Errorf("ytdirect.GetChannel: %w", err)
  }

  channelID := doc.Find("meta[itemprop=channelId]").AttrOr("content", "")
  channelTitle := doc.Find("meta[property='og:title']").AttrOr("content", "")

  ch := &Channel{
    ID:    channelID,
    Title: channelTitle,
  }

  for _, node := range doc.Find("script").Nodes {
    if node.FirstChild == nil || node.FirstChild.Type != html.TextNode {
      continue
    }

    jsContent := node.FirstChild.Data

    if !strings.HasPrefix(jsContent, "var ytInitialData =") {
      continue
    }

    jsContent = strings.TrimPrefix(jsContent, "var ytInitialData =")
    jsContent = strings.TrimSuffix(jsContent, ";")

    const (
      shelfListPath             = "contents.twoColumnBrowseResultsRenderer.tabs.0.tabRenderer.content.sectionListRenderer.contents"
      shelfTitlePath            = "itemSectionRenderer.contents.0.shelfRenderer.title.runs.0.text"
      playlistListPath          = "itemSectionRenderer.contents.0.shelfRenderer.content.horizontalListRenderer.items"
      playlistIDPath            = "gridPlaylistRenderer.playlistId"
      playlistChannelIDPath     = "gridPlaylistRenderer.longBylineText.runs.0.navigationEndpoint.browseEndpoint.browseId"
      playlistTitlePath         = "gridPlaylistRenderer.title.runs.0.text"
      playlistPublishedTimePath = "gridPlaylistRenderer.publishedTimeText.simpleText"
      playlistVideoCountPath    = "gridPlaylistRenderer.videoCountShortText.simpleText"
    )

    j, err := gabs.ParseJSON([]byte(jsContent))
    if err != nil {
      return nil, fmt.Errorf("ytdirect.GetChannel: %w", err)
    }

    for _, shelf := range j.Path(shelfListPath).Children() {
      if !shelf.ExistsP(shelfTitlePath) {
        continue
      }
      shelfTitle := shelf.Path(shelfTitlePath).Data().(string)

      var playlists []ChannelPlaylist

      for _, playlist := range shelf.Path(playlistListPath).Children() {
        if !playlist.ExistsP(playlistIDPath) || !playlist.ExistsP(playlistChannelIDPath) || !playlist.ExistsP(playlistTitlePath) {
          continue
        }

        playlistID := playlist.Path(playlistIDPath).Data().(string)
        playlistChannelID := playlist.Path(playlistChannelIDPath).Data().(string)
        playlistTitle := playlist.Path(playlistTitlePath).Data().(string)
        var playlistPublishedTime string
        if playlist.ExistsP(playlistPublishedTimePath) {
          playlistPublishedTime = playlist.Path(playlistPublishedTimePath).Data().(string)
        }
        var playlistVideoCount string
        if playlist.ExistsP(playlistVideoCountPath) {
          playlistVideoCount = playlist.Path(playlistVideoCountPath).Data().(string)
        }

        playlists = append(playlists, ChannelPlaylist{
          ID:            playlistID,
          ChannelID:     playlistChannelID,
          Title:         playlistTitle,
          PublishedTime: playlistPublishedTime,
          VideoCount:    playlistVideoCount,
        })
      }

      ch.Shelves = append(ch.Shelves, ChannelShelf{Title: shelfTitle, Playlists: playlists})
    }
  }

  return ch, nil
}

type Playlist struct {
  ID        string
  ChannelID string
  Title     string
  VideoIDs  []string
}

func GetPlaylist(ctx context.Context, id string) (*Playlist, error) {
  doc, err := getDocument(ctx, "https://www.youtube.com/playlist?list="+id)
  if err != nil {
    return nil, fmt.Errorf("ytdirect.GetPlaylist: %w", err)
  }

  var p Playlist

  for _, node := range doc.Find("script").Nodes {
    if node.FirstChild == nil || node.FirstChild.Type != html.TextNode {
      continue
    }

    jsContent := node.FirstChild.Data

    if !strings.HasPrefix(jsContent, "var ytInitialData =") {
      continue
    }

    jsContent = strings.TrimPrefix(jsContent, "var ytInitialData =")
    jsContent = strings.TrimSuffix(jsContent, ";")

    var (
      idPaths = []string{
        "header.playlistHeaderRenderer.playButton.buttonRenderer.navigationEndpoint.watchEndpoint.playlistId",
        "header.playlistHeaderRenderer.playlistHeaderBanner.heroPlaylistThumbnailRenderer.onTap.watchEndpoint.playlistId",
        "header.playlistHeaderRenderer.playlistId",
        "header.playlistHeaderRenderer.shufflePlayButton.buttonRenderer.navigationEndpoint.watchEndpoint.playlistId",
        "sidebar.playlistSidebarRenderer.items.0.playlistSidebarPrimaryInfoRenderer.navigationEndpoint.watchEndpoint.playlistId",
      }
      titlePaths = []string{
        "header.playlistHeaderRenderer.title.simpleText",
        "metadata.playlistMetadataRenderer.albumName",
        "metadata.playlistMetadataRenderer.title",
        "microformat.microformatDataRenderer.title",
      }
      entryListPath = "contents.twoColumnBrowseResultsRenderer.tabs.0.tabRenderer.content.sectionListRenderer.contents.0.itemSectionRenderer.contents.0.playlistVideoListRenderer.contents"
      channelIDPath = "playlistVideoRenderer.shortBylineText.runs.0.navigationEndpoint.browseEndpoint.browseId"
      videoIDPath   = "playlistVideoRenderer.videoId"
    )

    j, err := gabs.ParseJSON([]byte(jsContent))
    if err != nil {
      return nil, fmt.Errorf("ytdirect.GetPlaylist: %w", err)
    }

    for _, path := range idPaths {
      if j.ExistsP(path) {
        p.ID = j.Path(path).Data().(string)
        break
      }
    }

    for _, path := range titlePaths {
      if j.ExistsP(path) {
        p.Title = j.Path(path).Data().(string)
        break
      }
    }

    if j.ExistsP(entryListPath) {
      count, err := j.ArrayCountP(entryListPath)
      if err != nil {
        return nil, fmt.Errorf("ytdirect.GetPlaylist: could not get number of entries: %w", err)
      }

      for i := 0; i < count; i++ {
        element, err := j.ArrayElementP(i, entryListPath)
        if err != nil {
          return nil, fmt.Errorf("ytdirect.GetPlaylist: could not get entry %d: %w", i, err)
        }

        if element.ExistsP(channelIDPath) {
          if channelID, ok := element.Path(channelIDPath).Data().(string); ok {
            p.ChannelID = channelID
          }
        }

        if element.ExistsP(videoIDPath) {
          videoID, ok := element.Path(videoIDPath).Data().(string)
          if !ok {
            return nil, fmt.Errorf("ytdirect.GetPlaylist: could not get video id for entry %d", i)
          }

          p.VideoIDs = append(p.VideoIDs, videoID)
        }
      }
    }
  }

  if p.ID == "" {
    return nil, fmt.Errorf("ytdirect.GetPlaylist: could not find suitable data in page")
  }

  return &p, nil
}

type Video struct {
  ID          string
  ChannelID   string
  Title       string
  Description string
  PublishDate string
  UploadDate  string
}

func GetVideo(ctx context.Context, id string) (*Video, error) {
  doc, err := getDocument(ctx, "https://www.youtube.com/watch?v="+id)
  if err != nil {
    return nil, fmt.Errorf("ytdirect.GetVideo: %w", err)
  }

  var v Video

  for _, node := range doc.Find("script").Nodes {
    if node.FirstChild == nil || node.FirstChild.Type != html.TextNode {
      continue
    }

    jsContent := node.FirstChild.Data

    if !strings.HasPrefix(jsContent, "var ytInitialPlayerResponse =") {
      continue
    }

    jsContent = strings.TrimPrefix(jsContent, "var ytInitialPlayerResponse =")
    jsContent = strings.TrimSuffix(jsContent, ";")

    const (
      videoIDPath          = "videoDetails.videoId"
      videoChannelIDPath   = "videoDetails.channelId"
      videoTitlePath       = "microformat.playerMicroformatRenderer.title.simpleText"
      videoDescriptionPath = "microformat.playerMicroformatRenderer.description.simpleText"
      videoPublishDatePath = "microformat.playerMicroformatRenderer.publishDate"
      videoUploadDatePath  = "microformat.playerMicroformatRenderer.uploadDate"
    )

    j, err := gabs.ParseJSON([]byte(jsContent))
    if err != nil {
      return nil, fmt.Errorf("ytdirect.GetVideo: %w", err)
    }

    if j.ExistsP(videoIDPath) {
      v.ID = j.Path(videoIDPath).Data().(string)
    }
    if j.ExistsP(videoChannelIDPath) {
      v.ChannelID = j.Path(videoChannelIDPath).Data().(string)
    }
    if j.ExistsP(videoTitlePath) {
      v.Title = j.Path(videoTitlePath).Data().(string)
    }
    if j.ExistsP(videoDescriptionPath) {
      v.Description = j.Path(videoDescriptionPath).Data().(string)
    }
    if j.ExistsP(videoPublishDatePath) {
      v.PublishDate = j.Path(videoPublishDatePath).Data().(string)
    }
    if j.ExistsP(videoUploadDatePath) {
      v.UploadDate = j.Path(videoUploadDatePath).Data().(string)
    }
  }

  if v.ID == "" {
    return nil, fmt.Errorf("ytdirect.GetVideo: could not find suitable data in page")
  }

  return &v, nil
}
