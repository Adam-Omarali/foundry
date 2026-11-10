package youtube

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

// ErrNotVideoURL means the URL has no YouTube video ID in it.
var ErrNotVideoURL = errors.New("not a YouTube video URL")

// VideoID extracts the ID from youtube.com/watch?v=ID or youtu.be/ID.
func VideoID(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", ErrNotVideoURL
	}
	var id string
	switch {
	case u.Host == "youtu.be":
		id = strings.TrimPrefix(u.Path, "/")
	case strings.HasSuffix(u.Host, "youtube.com"):
		id = u.Query().Get("v")
	}
	if id == "" {
		return "", ErrNotVideoURL
	}
	return id, nil
}

func GetVideoDescription(ctx context.Context, videoID string) (string, error) {
	apiKey := os.Getenv("YOUTUBE_API_KEY")

	service, err := youtube.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return "", fmt.Errorf("creating YouTube client: %w", err)
	}

	resp, err := service.Videos.List([]string{"snippet"}).Id(videoID).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("fetching video %s: %w", videoID, err)
	}
	if len(resp.Items) == 0 {
		return "", fmt.Errorf("video %s not found", videoID)
	}
	return resp.Items[0].Snippet.Description, nil
}
