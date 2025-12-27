package converter

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kkdai/youtube/v2"
	"go.leapkit.dev/core/server"
)

func Convert(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		server.Errorf(w, http.StatusBadRequest, "failed to parse form: %w", err)
		return
	}

	videoURL := r.FormValue("youtube-url")
	if videoURL == "" {
		server.Errorf(w, http.StatusBadRequest, "youtube-url is required")
		return
	}

	outputPath := "./downloads"
	if err := os.MkdirAll(outputPath, os.ModePerm); err != nil {
		server.Errorf(w, http.StatusInternalServerError, "failed to create output directory: %w", err)
		return
	}

	result, err := YouTubeToMP3(videoURL, outputPath)
	if err != nil {
		server.Errorf(w, http.StatusInternalServerError, "conversion failed: %w", err)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+result.Name)
	w.Header().Set("Content-Type", "audio/mpeg")

	outputFile, err := os.Open(result.Path)
	if err != nil {
		server.Errorf(w, http.StatusInternalServerError, "error opening file: %w", err)
		return
	}

	defer outputFile.Close()

	if _, err := io.Copy(w, outputFile); err != nil {
		server.Errorf(w, http.StatusInternalServerError, "error sending file: %w", err)
		return
	}

	go os.Remove(result.Path)
}

// ConversionResult holds the result of the conversion
type ConversionResult struct {
	Name string
	Path string
}

// YouTubeToMP3 downloads a YouTube video and converts it to MP3
// using ffmpeg. It returns the path to the converted MP3 file.
// Requires ffmpeg to be installed and accessible in the system PATH.
func YouTubeToMP3(videoURL, outputPath string) (ConversionResult, error) {
	client := youtube.Client{}
	video, err := client.GetVideo(videoURL)
	if err != nil {
		return ConversionResult{}, fmt.Errorf("failed to get video info: %w", err)
	}

	formats := video.Formats.WithAudioChannels()
	if len(formats) == 0 {
		return ConversionResult{}, fmt.Errorf("no audio formats available")
	}

	var bestFormat *youtube.Format
	for i := range formats {
		if bestFormat == nil || formats[i].Bitrate > bestFormat.Bitrate {
			bestFormat = &formats[i]
		}
	}

	tempVideoFile := filepath.Join(os.TempDir(), "temp_video."+bestFormat.MimeType[6:strings.Index(bestFormat.MimeType, ";")])
	defer os.Remove(tempVideoFile)

	stream, _, err := client.GetStream(video, bestFormat)
	if err != nil {
		return ConversionResult{}, fmt.Errorf("failed to get video stream: %w", err)
	}
	defer stream.Close()

	file, err := os.Create(tempVideoFile)
	if err != nil {
		return ConversionResult{}, fmt.Errorf("failed to create temp file: %w", err)
	}

	_, err = file.ReadFrom(stream)
	file.Close()
	if err != nil {
		return ConversionResult{}, fmt.Errorf("failed to save video to temp file: %w", err)
	}

	sanitizedTitle := sanitizeFilename(video.Title)
	mp3Path := filepath.Join(outputPath, sanitizedTitle+".mp3")

	cmd := exec.Command("ffmpeg", "-i", tempVideoFile, "-vn", "-ar", "44100", "-ac", "2", "-b:a", "192k", mp3Path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return ConversionResult{}, fmt.Errorf("ffmpeg conversion failed: %w", err)
	}

	return ConversionResult{
		Name: sanitizedTitle + ".mp3",
		Path: mp3Path,
	}, nil
}

// sanitizeFilename removes invalid characters from filename
func sanitizeFilename(name string) string {
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalid {
		name = strings.ReplaceAll(name, char, "_")
	}
	return name
}
