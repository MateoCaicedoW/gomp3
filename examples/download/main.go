package main

import (
	"context"
	"fmt"
	"os"

	"github.com/MateoCaicedoW/gomp3/internal/system/services/mp3"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: download <youtube-url>")
		os.Exit(1)
	}

	videoURL := os.Args[1]

	// Create a new service
	svc := mp3.New()

	// Get video info
	info, err := svc.GetVideoInfo(videoURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting video info: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Downloading: %s by %s\n", info.Title, info.Author)
	fmt.Printf("Duration: %s\n", info.Duration)

	// Create output file
	safeName := mp3.SanitizeFilename(info.Title) + ".mp3"
	file, err := os.Create(safeName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	fmt.Printf("Saving to: %s\n", safeName)

	// Download and convert
	opts := mp3.DefaultOptions()
	if err := svc.ConvertToWriter(context.Background(), videoURL, file, opts); err != nil {
		fmt.Fprintf(os.Stderr, "Error converting: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Download complete!")
}
