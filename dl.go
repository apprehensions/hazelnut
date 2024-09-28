package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	bc "github.com/apprehensions/hazelnut/bandcamp"
	"github.com/cheggaaa/pb/v3"
	"golang.org/x/sync/semaphore"
)

var filenameSanitizer = strings.NewReplacer(
	"/", "_",
	"?", "_",
	"<", "_",
	">", "_",
	":", "_",
	"|", "_",
	"\"", "_",
	"*", "_",
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

var NoBar pb.ProgressBarTemplate = `{{with string . "prefix"}}{{.}} {{end}}{{counters . }} {{with string . "suffix"}}{{.}}{{end}}`

type CollectionDownloader struct {
	pool *pb.Pool
	root *pb.ProgressBar

	dir    string
	format string
	ext    string

	*bc.Client
}

func New(dir string, format string, client *bc.Client) (*CollectionDownloader, error) {
	root := NoBar.New(0)
	pool := pb.NewPool(root)

	ext, ok := formatExtension[format]
	if !ok {
		return nil, fmt.Errorf("bad format requested: %s", format)
	}

	return &CollectionDownloader{
		pool:   pool,
		root:   root,
		dir:    dir,
		format: format,
		ext:    ext,
		Client: client,
	}, nil
}

func (cd *CollectionDownloader) Start(ctx context.Context) error {
	if err := cd.pool.Start(); err != nil {
		return fmt.Errorf("progress pool: %w", err)
	}

	items, err := cd.Items()
	if err != nil {
		return err
	}

	max := int64(6)
	sem := semaphore.NewWeighted(max)
	ctx, cancel := context.WithCancelCause(ctx)
	_ = cancel

	for _, item := range items {
		_ = item
		if err := sem.Acquire(ctx, 1); err != nil {
			break
		}
		go func() {
			defer func() {
				sem.Release(1)
				cd.root.Increment()
			}()
			if err := cd.Download(&item); err != nil {
				cancel(err)
			}
		}()
	}

	err = sem.Acquire(ctx, max)
	cd.Cleanup()
	if err != nil && errors.Is(err, context.Canceled) {
		return context.Cause(ctx)
	}
	return err
}

func (cd *CollectionDownloader) Cleanup() {
	cleanup, _ := filepath.Glob(filepath.Join(cd.dir, "*", "*", "*_hazelnut.zip"))
	for _, archive := range cleanup {
		os.Remove(archive)
	}
}

func (cd *CollectionDownloader) Items() ([]bc.Item, error) {
	fs, err := cd.GetCollectionSummary()
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	cd.root.Set("prefix", fs.Collection.Username)

	items, err := cd.GetItems(fs.ID)
	if err != nil {
		return nil, fmt.Errorf("get items: %w", err)
	}
	cd.root.SetTotal(int64(len(items)))

	return items, err
}

func (cd *CollectionDownloader) DownloadPath(i *bc.Item) string {
	band := filenameSanitizer.Replace(i.BandName)
	title := filenameSanitizer.Replace(i.Title)

	out := filepath.Join(cd.dir, band, title)
	if i.Type.IsTrack() {
		out += cd.ext
	} else {
		out = filepath.Join(out, fmt.Sprint(i.ID)+"_hazelnut.zip")
	}

	var err error
	if i.Type.IsTrack() {
		_, err = os.Stat(out)
	} else {
		_, err = os.Stat(filepath.Dir(out))
	}
	if err == nil {
		return ""
	}

	return out
}

func (cd *CollectionDownloader) Download(i *bc.Item) error {
	name := cd.DownloadPath(i)
	if name == "" {
		return nil // skip, already downloaded
	}
	dir := filepath.Dir(name)

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create dirs: %w", err)
	}

	bar := pb.New(0)
	bar.SetTemplate(pb.Simple)
	bar.Set("prefix", fmt.Sprintf("%s%d", i.Type.Short(), i.ID))
	bar.Set(pb.CleanOnFinish, true)
	cd.pool.Add(bar)
	defer bar.Finish()

	dl, err := cd.GetDigitalItemDownload(i, cd.format)
	if err != nil {
		return fmt.Errorf("get digital item: %w", err)
	}

	comp, err := os.Create(name)
	if err != nil {
		return fmt.Errorf("download create: %w", err)
	}
	defer func() {
		comp.Close()
		if i.Type.IsAlbum() {
			os.Remove(comp.Name())
		}
	}()

	resp, err := cd.Client.Download(dl)
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	if resp.ContentLength < 0 {
		return errors.New("Content-Length missing")
	}

	bar.SetTotal(resp.ContentLength)
	rd := bar.NewProxyReader(resp.Body)

	if _, err := io.Copy(comp, rd); err != nil {
		return fmt.Errorf("download copy: %w", err)
	}

	if i.Type.IsTrack() {
		return nil
	}

	if err := cd.Extract(i, comp, dir); err != nil {
		return fmt.Errorf("extract: %w", err)
	}

	return nil
}
