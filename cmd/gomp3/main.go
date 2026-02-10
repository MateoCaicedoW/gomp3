package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/MateoCaicedoW/gomp3/internal/system/services/mp3"
)

func main() {
	var (
		output     = flag.String("o", "", "Output filename (default: video title)")
		bitrate    = flag.String("b", "64k", "Audio bitrate (e.g., 64k, 128k, 192k)")
		sampleRate = flag.Int("r", 22050, "Sample rate in Hz (e.g., 22050, 44100)")
		channels   = flag.Int("c", 1, "Audio channels: 1 for mono, 2 for stereo")
		infoOnly   = flag.Bool("i", false, "Show video info only, don't download")
	)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <youtube-url>\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Download YouTube videos as MP3 audio files.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s https://youtube.com/watch?v=...\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -o mysong.mp3 https://youtube.com/watch?v=...\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -b 128k -c 2 https://youtube.com/watch?v=...\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -i https://youtube.com/watch?v=...\n", os.Args[0])
	}
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	videoURL := flag.Arg(0)

	svc := mp3.New()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nInterrupted, cancelling...")
		cancel()
	}()

	info, err := svc.GetVideoInfo(videoURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting video info: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Title:    %s\n", info.Title)
	fmt.Printf("Author:   %s\n", info.Author)
	fmt.Printf("Duration: %s\n", info.Duration)

	if *infoOnly {
		return
	}

	filename := *output
	if filename == "" {
		filename = mp3.SanitizeFilename(info.Title) + ".mp3"
	}

	if _, err := os.Stat(filename); err == nil {
		fmt.Fprintf(os.Stderr, "Error: file '%s' already exists\n", filename)
		os.Exit(1)
	}

	file, err := os.Create(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	fmt.Printf("Output:   %s\n", filename)
	fmt.Printf("Bitrate:  %s, Sample Rate: %d Hz, Channels: %d\n", *bitrate, *sampleRate, *channels)
	fmt.Println("Downloading...")

	opts := &mp3.Options{
		SampleRate: *sampleRate,
		Channels:   *channels,
		Bitrate:    *bitrate,
		Format:     "mp3",
	}

	if err := svc.ConvertToWriter(ctx, videoURL, file, opts); err != nil {
		os.Remove(filename)
		fmt.Fprintf(os.Stderr, "Error downloading: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Done!")
}
