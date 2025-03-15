package main

import (
	"archive/zip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func ExtractAlbum(src *os.File, dir string) error {
	s, err := src.Stat()
	if err != nil {
		return err
	}

	r, err := zip.NewReader(src, s.Size())
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			return errors.New("unexpected directory")
		}

		// life - demo two - 05 twelve travel.flac -> 05 twelve travel.flac
		track := strings.Split(f.Name, " - ")
		name := filepath.Join(dir, track[len(track)-1])
		if err := unzipFile(f, name); err != nil {
			return err
		}
	}

	return nil
}

func unzipFile(src *zip.File, name string) error {
	f, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, src.Mode())
	if err != nil {
		return err
	}
	defer f.Close()

	z, err := src.Open()
	if err != nil {
		return err
	}
	defer z.Close()

	if _, err := io.Copy(f, z); err != nil {
		return err
	}

	return nil
}
