// Package mp3 provides a service for downloading and converting YouTube videos to MP3 format.
// It supports streaming conversion directly to an io.Writer, making it suitable for web servers,
// file downloads, and other applications.
package mp3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/kkdai/youtube/v2"
)

// VideoInfo contains metadata about a YouTube video.
type VideoInfo struct {
	Title    string
	Author   string
	Duration string
	VideoID  string
}

// Service provides methods for downloading and converting YouTube videos.
type Service struct {
	client youtube.Client
}

// New creates a new Service instance.
func New() *Service {
	return &Service{
		client: youtube.Client{},
	}
}

// GetVideoInfo retrieves metadata about a YouTube video without downloading it.
func (s *Service) GetVideoInfo(videoURL string) (*VideoInfo, error) {
	video, err := s.client.GetVideo(videoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get video info: %w", err)
	}

	return &VideoInfo{
		Title:    video.Title,
		Author:   video.Author,
		Duration: video.Duration.String(),
		VideoID:  video.ID,
	}, nil
}

// ConvertToWriter downloads a YouTube video and converts it to MP3,
// streaming the output directly to the provided io.Writer.
// The videoURL can be a full YouTube URL or video ID.
// If opts is nil, DefaultOptions() will be used.
func (s *Service) ConvertToWriter(ctx context.Context, videoURL string, w io.Writer, opts *Options) error {
	if w == nil {
		return fmt.Errorf("writer is required")
	}

	resolved := normalizeOptions(opts)

	video, err := s.client.GetVideo(videoURL)
	if err != nil {
		return fmt.Errorf("failed to get video info: %w", err)
	}

	format := s.selectBestAudioFormat(video)
	if format == nil {
		return fmt.Errorf("no audio formats available")
	}

	stream, _, err := s.client.GetStream(video, format)
	if err != nil {
		return fmt.Errorf("failed to get video stream: %w", err)
	}
	defer stream.Close()

	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-i", "pipe:0",
		"-vn",
		"-ar", fmt.Sprintf("%d", resolved.SampleRate),
		"-ac", fmt.Sprintf("%d", resolved.Channels),
		"-b:a", resolved.Bitrate,
		"-f", resolved.Format,
		"pipe:1",
	)

	cmd.Stdin = stream
	cmd.Stdout = w
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg conversion failed: %w", err)
	}

	return nil
}

// Convert downloads a YouTube video, converts it to MP3, and returns the audio data.
// This is a convenience method that buffers the output in memory.
// For large files or server applications, use ConvertToWriter instead.
func (s *Service) Convert(ctx context.Context, videoURL string, opts *Options) ([]byte, error) {
	var buf bytes.Buffer
	if err := s.ConvertToWriter(ctx, videoURL, &buf, opts); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (s *Service) selectBestAudioFormat(video *youtube.Video) *youtube.Format {
	formats := video.Formats.Type("audio")
	if len(formats) == 0 {
		formats = video.Formats.WithAudioChannels()
	}

	if len(formats) == 0 {
		return nil
	}

	var bestFormat *youtube.Format
	for i := range formats {
		f := &formats[i]
		if f.QualityLabel == "" {
			if bestFormat == nil || f.Bitrate < bestFormat.Bitrate {
				bestFormat = f
			}
		}
	}

	if bestFormat == nil {
		bestFormat = &formats[0]
	}

	return bestFormat
}

func normalizeOptions(opts *Options) Options {
	resolved := DefaultOptions()
	if opts == nil {
		return *resolved
	}

	if opts.SampleRate != 0 {
		resolved.SampleRate = opts.SampleRate
	}
	if opts.Channels != 0 {
		resolved.Channels = opts.Channels
	}
	if opts.Bitrate != "" {
		resolved.Bitrate = opts.Bitrate
	}
	if opts.Format != "" {
		resolved.Format = opts.Format
	}

	return *resolved
}

// SanitizeFilename removes invalid characters from a filename.
func SanitizeFilename(name string) string {
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	result := name
	for _, char := range invalid {
		result = strings.ReplaceAll(result, char, "_")
	}
	return result
}
