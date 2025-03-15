package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
)

var (
	dir     string
	jobs    int64
	format  string
	cookies string
)

func init() {
	flag.StringVar(&format, "format", "flac", "audio format to use")
	flag.Int64Var(&jobs, "jobs", 6, "amount of parallel jobs to use to download")
	flag.StringVar(&dir, "o", "collection", "directory to download albums to")
	flag.StringVar(&cookies, "cookies", "hazelnut-cookies.txt", "bandcamp user cookies file path")
}

func main() {
	flag.Parse()

	c, err := os.ReadFile(cookies)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(22)
	}

	if c[len(c)-1] == '\n' {
		c = c[:len(c)-1]
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	downloader, err := new(ctx, string(c))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	err = downloader.downloadCollection()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(11)
	}
}
