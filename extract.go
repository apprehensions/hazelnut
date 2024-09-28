package main

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"

	bc "github.com/apprehensions/hazelnut/bandcamp"
)

// life - demo two - 05 twelve travel.flac -> 05 twelve travel.flac
var TrackPattern = regexp.MustCompile(`^[^-]* - [^-]* - (\d+.*\.\w+)$`)

func (cd *CollectionDownloader) Extract(i *bc.Item, src *os.File, dir string) error {
	if _, err := src.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("seek: %w", err)
	}

	st, err := src.Stat()
	if err != nil {
		return fmt.Errorf("stat: %w", err)
	}

	r, err := zip.NewReader(src, st.Size())
	if err != nil {
		return fmt.Errorf("zip reader: %w", err)
	}

	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			return errors.New("unexpected directory")
		}

		name := f.Name
		if filepath.Ext(name) == cd.ext {
			match := TrackPattern.FindStringSubmatch(name)
			if len(match) != 2 {
				return fmt.Errorf("unexpected track filename format: %s", name)
			}
			name = match[1]
		}
		name = filepath.Join(dir, name)

		if err := unzipFile(f, name); err != nil {
			return fmt.Errorf("unzip file: %w")
		}
	}

	return nil
}

func unzipFile(src *zip.File, name string) error {
	f, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, src.Mode())
	if err != nil {
		return fmt.Errorf("create: %w", err)
	}
	defer f.Close()

	z, err := src.Open()
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	defer z.Close()

	if _, err := io.Copy(f, z); err != nil {
		return fmt.Errorf("copy: %w", err)
	}

	return nil
}
