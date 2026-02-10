GoMP3
=====

GoMP3 is a tiny YouTube-to-MP3 converter written in Go. It provides both a web interface and a command-line tool for downloading YouTube videos as MP3 audio files.

Features
- **CLI Tool** - Download YouTube videos as MP3 from the command line
- **Web App** - Browser-based interface with responsive design and dark mode
- Streams converted audio directly (no temp files)
- Ships with Tailwind-based styles and Docker support with ffmpeg preinstalled
- **Internal MP3 Service** for processing conversions

Requirements
- Go 1.24+
- ffmpeg available on your PATH (for local runs)
- Tailwind CLI helper `tailo` for rebuilding CSS in development
	- Install with `go tool tailo download`

Installation
------------

### Install the CLI Tool
```bash
go install github.com/MateoCaicedoW/gomp3/cmd/gomp3@latest
```

CLI Usage
---------

The `gomp3` command-line tool allows you to download YouTube videos as MP3 files:

```bash
# Basic usage
gomp3 https://youtube.com/watch?v=...

# Specify output filename
gomp3 -o mysong.mp3 https://youtube.com/watch?v=...

# Higher quality (stereo, 128k bitrate, 44.1kHz)
gomp3 -b 128k -c 2 -r 44100 https://youtube.com/watch?v=...

# Show video info only
gomp3 -i https://youtube.com/watch?v=...
```

### CLI Options
```
-b string
    Audio bitrate (e.g., 64k, 128k, 192k) (default "64k")
-c int
    Audio channels: 1 for mono, 2 for stereo (default 1)
-i  Show video info only, don't download
-o string
    Output filename (default: video title)
-r int
    Sample rate in Hz (e.g., 22050, 44100) (default 22050)
```

Web App Usage
-------------

### Run the Web Server Locally
1) Install deps: `go mod download`
2) Ensure ffmpeg is installed (macOS: `brew install ffmpeg`)
3) Run the app: `go tool dev --watch.extensions=.go,.css,.js `
4) Visit http://localhost:3000 and paste a YouTube link

### Docker
- Build: `docker build -t gomp3 .`
- Run: `docker run --rm -p 3000:3000 gomp3`
- ffmpeg is included in the image

### Configuration
- `HOST` (default `0.0.0.0`)
- `PORT` (default `3000`)
- `SESSION_SECRET` (default random string)
- `SESSION_NAME` (default `leapkit_session`)

Project Structure
-----------------

```
cmd/
  gomp3/                    # CLI tool for downloading MP3s
  app/                      # Web server application
internal/
  converter/                # HTTP handlers for web interface
  system/
    services/mp3/          # MP3 conversion service
    assets/                # Static assets
    helpers/               # Utility helpers
examples/                   # Example applications
```

Internal MP3 Service
--------------------

The `internal/system/services/mp3` package provides the core YouTube to MP3 conversion functionality:

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/MateoCaicedoW/gomp3/internal/system/services/mp3"
)

func main() {
    // Create a new service
    svc := mp3.New()

    // Get video info
    info, err := svc.GetVideoInfo("https://youtube.com/watch?v=...")
    if err != nil {
        panic(err)
    }
    fmt.Printf("Title: %s\n", info.Title)

    // Create output file
    filename := mp3.SanitizeFilename(info.Title) + ".mp3"
    file, err := os.Create(filename)
    if err != nil {
        panic(err)
    }
    defer file.Close()

    // Download and convert with default options
    err = svc.ConvertToWriter(context.Background(), "https://youtube.com/watch?v=...", file, nil)
    if err != nil {
        panic(err)
    }
}
```

### Advanced Options

You can customize the conversion parameters:

```go
opts := &mp3.Options{
    SampleRate: 44100,  // Higher quality (default: 22050)
    Channels:   2,      // Stereo (default: 1 - mono)
    Bitrate:    "128k", // Higher bitrate (default: "64k")
    Format:     "mp3",  // Output format (default: "mp3")
}

err := svc.ConvertToWriter(ctx, videoURL, writer, opts)
```

See the `examples/` directory for complete working examples.

Notes
-----
- The converter picks the best available audio stream from YouTube, runs ffmpeg, then streams the result
- The MP3 service is located in `internal/system/services/mp3/`
- When using the service, ensure ffmpeg is installed and available in your PATH
- The CLI tool supports signal handling (Ctrl+C to cancel downloads gracefully)
