package main

import (
	"flag"
	"log"
	"log/slog"
	"os"

	bc "github.com/apprehensions/hazelnut/bandcamp"
	"github.com/lmittmann/tint"
)

var (
	outputDir string
	format    string
	cookies   string
	fileExt   string
)

var formatExtension = map[string]string{
	"mp3-v0":        ".mp3",
	"mp3-320":       ".mp3",
	"flac":          ".flac",
	"aac-hi":        ".m4a",
	"vorbis":        ".ogg",
	"alac":          ".m4a",
	"wav":           ".wav",
	"aiff-lossless": ".aiff",
}

func init() {
	flag.StringVar(&format, "format", "flac", "audio format to use")
	flag.StringVar(&outputDir, "o", ".", "directory to download albums to")
	flag.StringVar(&cookies, "cookies", "hazelnut-cookies.txt", "bandcamp user cookies file path")
}

func main() {
	flag.Parse()

	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level: slog.LevelDebug,
		}),
	))

	ext, ok := formatExtension[format]
	if !ok {
		log.Fatalf("unhandled audio format extension delegation: %s", format)
	}
	fileExt = ext

	c, err := os.ReadFile(cookies)
	if err != nil {
		log.Fatal(err)
	}

	if c[len(c)-1] == '\n' {
		c = c[:len(c)-1]
	}

	d := &downloader{c: bc.New(string(c))}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		log.Fatal(err)
	}

	err = d.DownloadAll()
	if err != nil {
		log.Fatal(err)
	}
}
