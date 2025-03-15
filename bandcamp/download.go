package bandcamp

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/net/html"
)

var Extensions = map[string]string{
	"mp3-v0":        ".mp3",
	"mp3-320":       ".mp3",
	"flac":          ".flac",
	"aac-hi":        ".m4a",
	"vorbis":        ".ogg",
	"alac":          ".m4a",
	"wav":           ".wav",
	"aiff-lossless": ".aiff",
}

type RedownloadItem struct {
	Item
	URL string
}

type DownloadItem struct {
	Item
	Release ItemRelease
	Download
}

type Download struct {
	Size     string `json:"size_mb"`
	Encoding string `json:"encoding_name"`
	URL      string `json:"url"`
}

type DigitalItem struct {
	Downloads ItemRedownloads `json:"downloads"`
	Release   ItemRelease     `json:"release_date"`
}

// map[Extensions.key]Download
type ItemRedownloads map[string]Download

func (c *Client) GetRedownloadItems(id FanID) ([]RedownloadItem, error) {
	var items []RedownloadItem

	ci, err := c.GetCollectionItems(id)
	if err != nil {
		return nil, err
	}

	for _, item := range ci.Items {
		name := fmt.Sprint(item.SaleItemType, item.SaleItemID)
		url, ok := ci.RedownloadURLs[name]
		if !ok {
			return nil, fmt.Errorf("missing download for %s", name)
		}

		items = append(items, RedownloadItem{
			Item: item, URL: url,
		})
	}

	return items, nil
}

func (c *Client) GetDownloadItem(item *RedownloadItem, format string) (*DownloadItem, error) {
	digital, err := c.GetItemRedownloads(item)
	if err != nil {
		return nil, err
	}

	for _, d := range digital.Downloads {
		if d.Encoding == format {
			return &DownloadItem{
				Item:     item.Item,
				Download: d,
				Release:  digital.Release,
			}, nil
		}
	}

	return nil, fmt.Errorf("format %s unavailable", format)
}

func (c *Client) GetItemRedownloads(item *RedownloadItem) (*DigitalItem, error) {
	var blob string
	dd := struct {
		Items []DigitalItem `json:"download_items"`
	}{}

	req, err := http.NewRequest("GET", item.URL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	z := html.NewTokenizer(resp.Body)
	for {
		t := z.Next()
		if t == html.ErrorToken {
			if z.Err() == io.EOF {
				break
			}

			return nil, z.Err()
		}

		k := z.Token()
		if k.Type != html.StartTagToken {
			continue
		}

		for _, a := range k.Attr {
			if a.Key == "data-blob" {
				blob = a.Val
				break
			}
		}

		if blob != "" {
			break
		}
	}

	if blob == "" {
		return nil, errors.New("no download data found")
	}

	err = json.Unmarshal([]byte(blob), &dd)
	if err != nil {
		return nil, err
	}

	return &dd.Items[0], nil
}
