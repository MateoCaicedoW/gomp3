// Package mp3 provides a service for downloading and converting YouTube videos to MP3 format.
// It supports streaming conversion directly to an io.Writer, making it suitable for web servers,
// file downloads, and other applications.
package mp3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

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

// New creates a new Service instance with custom HTTP client configuration.
func New() *Service {
	// Create HTTP client with proper timeout to avoid being blocked
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	return &Service{
		client: youtube.Client{
			HTTPClient: httpClient,
		},
	}
}

// GetVideoInfo retrieves metadata about a YouTube video without downloading it.
func (s *Service) GetVideoInfo(videoURL string) (*VideoInfo, error) {
	// Extract clean video URL without playlist parameters
	cleanURL := extractVideoURL(videoURL)

	video, err := s.client.GetVideo(cleanURL)
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

	// Extract clean video URL without playlist parameters
	cleanURL := extractVideoURL(videoURL)

	// First try using yt-dlp if available (more reliable)
	if err := s.convertWithYTDLP(ctx, cleanURL, w, opts); err == nil {
		return nil
	}

	// Fallback to kkdai/youtube library
	return s.convertWithLibrary(ctx, cleanURL, w, opts)
}

// extractVideoURL extracts the video ID from various YouTube URL formats
// and returns a clean URL with just the video ID
func extractVideoURL(videoURL string) string {
	// If it's already just an ID (11 characters, alphanumeric, _, -)
	if matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]{11}$`, videoURL); matched {
		return "https://www.youtube.com/watch?v=" + videoURL
	}

	// Parse the URL
	u, err := url.Parse(videoURL)
	if err != nil {
		return videoURL
	}

	// Extract video ID from query parameters
	videoID := u.Query().Get("v")
	if videoID == "" {
		return videoURL
	}

	// Return clean URL with just the video ID
	return "https://www.youtube.com/watch?v=" + videoID
}

func (s *Service) convertWithYTDLP(ctx context.Context, videoURL string, w io.Writer, opts *Options) error {
	resolved := normalizeOptions(opts)

	// Check if yt-dlp is available
	if _, err := exec.LookPath("yt-dlp"); err != nil {
		return fmt.Errorf("yt-dlp not found")
	}

	cmd := exec.CommandContext(ctx, "yt-dlp",
		"--no-warnings",
		"--quiet",
		"--no-playlist",
		"-f", "bestaudio[ext=m4a]/bestaudio",
		"-o", "-",
		"--ffmpeg-location", "ffmpeg",
		videoURL,
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start yt-dlp: %w", err)
	}

	// Convert the downloaded audio with ffmpeg
	ffmpegCmd := exec.CommandContext(ctx, "ffmpeg",
		"-hide_banner",
		"-loglevel", "error",
		"-i", "pipe:0",
		"-vn",
		"-ar", fmt.Sprintf("%d", resolved.SampleRate),
		"-ac", fmt.Sprintf("%d", resolved.Channels),
		"-b:a", resolved.Bitrate,
		"-f", resolved.Format,
		"-",
	)

	ffmpegCmd.Stdin = stdout
	ffmpegCmd.Stdout = w

	var stderr bytes.Buffer
	ffmpegCmd.Stderr = &stderr

	ffmpegErrChan := make(chan error, 1)
	go func() {
		ffmpegErrChan <- ffmpegCmd.Run()
	}()

	// Wait for both processes
	cmdErr := cmd.Wait()
	ffmpegErr := <-ffmpegErrChan

	if ffmpegErr != nil {
		errMsg := stderr.String()
		if errMsg != "" {
			return fmt.Errorf("ffmpeg conversion failed: %s", strings.TrimSpace(errMsg))
		}
		return fmt.Errorf("ffmpeg conversion failed: %w", ffmpegErr)
	}

	if cmdErr != nil {
		return fmt.Errorf("yt-dlp failed: %w", cmdErr)
	}

	return nil
}

func (s *Service) convertWithLibrary(ctx context.Context, videoURL string, w io.Writer, opts *Options) error {
	resolved := normalizeOptions(opts)

	video, err := s.client.GetVideo(videoURL)
	if err != nil {
		return fmt.Errorf("failed to get video info: %w", err)
	}

	format := s.selectBestAudioFormat(video)
	if format == nil {
		return fmt.Errorf("no audio formats available")
	}

	// YouTube blocks direct streaming, so we download to a temp file first
	stream, _, err := s.client.GetStream(video, format)
	if err != nil {
		return fmt.Errorf("failed to get video stream: %w", err)
	}
	defer stream.Close()

	// Download to temp file
	tempFile, err := os.CreateTemp("", "gomp3-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath)
	defer tempFile.Close()

	// Copy stream to temp file
	_, err = io.Copy(tempFile, stream)
	if err != nil {
		return fmt.Errorf("failed to download video (YouTube may be blocking this request). Try installing yt-dlp: brew install yt-dlp")
	}
	tempFile.Close()

	// Convert using ffmpeg
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-hide_banner",
		"-loglevel", "error",
		"-i", tempPath,
		"-vn",
		"-ar", fmt.Sprintf("%d", resolved.SampleRate),
		"-ac", fmt.Sprintf("%d", resolved.Channels),
		"-b:a", resolved.Bitrate,
		"-f", resolved.Format,
		"-",
	)

	cmd.Stdout = w

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errMsg := stderr.String()
		if errMsg != "" {
			return fmt.Errorf("ffmpeg conversion failed: %s", strings.TrimSpace(errMsg))
		}
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
