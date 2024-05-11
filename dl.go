package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	bc "github.com/apprehensions/hazelnut/bandcamp"
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

type Item struct {
	DownloadURL string
	*bc.CollectionItem
}

type downloader struct {
	c *bc.Client
}

func (d *downloader) DownloadAll() error {
	slog.Info("Retrieving collection summary")
	fs, err := d.c.GetCollectionSummary()
	if err != nil {
		return fmt.Errorf("get summary: %w", err)
	}

	slog.Info("Retrieved collection summary",
		"fan_id", fs.ID, "user", fs.Collection.Username)

	ci, err := d.c.GetCollectionItems(fs.ID, int(^uint(0)>>1))
	if err != nil {
		return fmt.Errorf("get collection: %w", err)
	}

	slog.Info("Retrieved collection items", "count", len(ci.Items))

	max := int64(6)
	sem := semaphore.NewWeighted(max)
	ctx := context.TODO()

	for si, u := range ci.RedownloadURLs {
		for _, i := range ci.Items {
			if si != fmt.Sprint(i.SaleItemType, i.SaleItemID) {
				continue
			}

			if err := sem.Acquire(ctx, 1); err != nil {
				break
			}

			go func() {
				defer sem.Release(1)

				i := &Item{
					CollectionItem: &i,
					DownloadURL:    u,
				}

				if err := d.Download(i); err != nil {
					slog.Error("Failed to download item", "item", i.String(), "error", err)
				}
			}()
		}
	}

	if err := sem.Acquire(ctx, max); err != nil {
		return fmt.Errorf("wait: %w", err)
	}

	slog.Info("Downloaded all items")

	return nil
}

func (d *downloader) FindDownload(i *Item) (*bc.DigitalItemDownload, error) {
	slog.Info("Fetching item for suitable format", "format", format, "item", i.String())

	di, err := d.c.GetDigitalItem(i.DownloadURL)
	if err != nil {
		return nil, err
	}

	for _, fdl := range di.Downloads {
		if fdl.Encoding == format {
			return &fdl, nil
		}
	}

	return nil, fmt.Errorf("format %s unavailable", format)
}

func (d *downloader) Download(i *Item) error {
	band := filenameSanitizer.Replace(i.BandName)
	title := filenameSanitizer.Replace(i.Title)

	out := filepath.Join(outputDir, band, title+fileExt)
	if i.Type.IsAlbum() {
		out = filepath.Join(outputDir, band, title)
	}

	_, err := os.Stat(out)
	if err == nil {
		slog.Warn("Skipping download, already exists", "item", i.String(), "path", out)
		return nil
	}

	dl, err := d.FindDownload(i)
	if err != nil {
		return err
	}

	tmp, err := os.CreateTemp("", strconv.FormatInt(int64(i.ID), 10)+".*")
	if err != nil {
		return err
	}
	defer func() {
		tmp.Close()
	}()

	slog.Info("Downloading", "item", i.String(),
		"format", dl.Encoding, "size", dl.Size, "tmp", tmp.Name())

	resp, err := d.c.Download(dl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	if resp.ContentLength < 0 {
		return errors.New("Content-Length missing")
	}

	if _, err := io.Copy(tmp, resp.Body); err != nil {
		return err
	}

	if _, err := tmp.Seek(0, io.SeekStart); err != nil {
		return err
	}

	slog.Info("Extracting", "item", i.String(), "path", out)

	if i.Type.IsAlbum() {
		return ExtractAlbum(tmp, out)
	}

	tmp.Close()
	return os.Rename(tmp.Name(), out)
}
