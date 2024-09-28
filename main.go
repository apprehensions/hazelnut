package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	bc "github.com/apprehensions/hazelnut/bandcamp"
)

func main() {
	aff := flag.String("format", "flac", "audio format to use")
	dir := flag.String("o", ".", "directory to download albums to")
	ckf := flag.String("cookies", "hazelnut-cookies.txt", "bandcamp user cookies file path")
	flag.Parse()

	c := client(*ckf)
	cd, err := New(*dir, *aff, c)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	err = cd.Start(ctx)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}

func client(name string) *bc.Client {
	ck, err := os.ReadFile(name)
	if err != nil {
		fmt.Println("cookies:", err)
		os.Exit(1)
	}

	if ck[len(ck)-1] == '\n' {
		ck = ck[:len(ck)-1]
	}

	return bc.New(string(ck))
}
