package mp3

// Options configures the MP3 conversion parameters.
type Options struct {
	// SampleRate is the audio sample rate in Hz (default: 22050)
	SampleRate int
	// Channels is the number of audio channels: 1 for mono, 2 for stereo (default: 1)
	Channels int
	// Bitrate is the audio bitrate (default: "64k")
	Bitrate string
	// Format is the output format, typically "mp3" (default: "mp3")
	Format string
}

// DefaultOptions returns the default conversion options.
func DefaultOptions() *Options {
	return &Options{
		SampleRate: 22050,
		Channels:   1,
		Bitrate:    "64k",
		Format:     "mp3",
	}
}
