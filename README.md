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
- **yt-dlp** (recommended for better compatibility with YouTube)
  - Install with: `brew install yt-dlp` (macOS) or see [yt-dlp installation](https://github.com/yt-dlp/yt-dlp#installation)
- Tailwind CLI helper `tailo` for rebuilding CSS in development
  - Install with `go tool tailo download`

Installation
------------

### Install the CLI Tool
```bash
go install github.com/MateoCaicedoW/gomp3/cmd/gomp3@latest
```

### Install the Web Server
```bash
go install github.com/MateoCaicedoW/gomp3/cmd/app@latest
```

### Optional: Install yt-dlp (Recommended)
```bash
# macOS
brew install yt-dlp

# Linux
sudo curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o /usr/local/bin/yt-dlp
sudo chmod a+rx /usr/local/bin/yt-dlp

# Windows (with scoop)
scoop install yt-dlp
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
2) Ensure ffmpeg and yt-dlp are installed
3) Run the app: `go tool dev --watch.extensions=.go,.css,.js `
4) Visit http://localhost:3000 and paste a YouTube link

### Docker
- Build: `docker build -t gomp3 .`
- Run: `docker run --rm -p 3000:3000 gomp3`
- ffmpeg is included in the image (add yt-dlp to Dockerfile for best results)

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

Notes
-----
- **yt-dlp is highly recommended** - YouTube blocks many automated downloads. Installing yt-dlp provides much better compatibility.
- Without yt-dlp, some videos may fail to download with "403 Forbidden" errors
- The converter picks the best available audio stream from YouTube, runs ffmpeg, then streams the result
- The MP3 service is located in `internal/system/services/mp3/`
- When using the service, ensure ffmpeg is installed and available in your PATH
- The CLI tool supports signal handling (Ctrl+C to cancel downloads gracefully)

Troubleshooting
---------------

### "Error downloading: failed to download video" or 403 errors
This means YouTube is blocking the download. Install yt-dlp for better compatibility:
```bash
brew install yt-dlp  # macOS
```

### ffmpeg not found
Install ffmpeg:
```bash
brew install ffmpeg  # macOS
sudo apt-get install ffmpeg  # Ubuntu/Debian
```

### Slow downloads
The conversion process downloads the video first, then converts it to MP3. For longer videos, this can take several minutes.
