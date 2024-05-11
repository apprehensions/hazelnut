package bandcamp

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"golang.org/x/net/html"
)

var userAgent = "Mozilla/5.0 (X11; Linux x86_64; rv:125.0) Gecko/20100101 Firefox/125.0"

var ErrNoData = errors.New("no download data found")

type DigitalItemDownload struct {
	Size        string `json:"size_mb"`
	Description string `json:"description"`
	Encoding    string `json:"encoding_name"`
	URL         string `json:"url"`
}

type DigitalItem struct {
	// Many fields omitted - out of scope
	Type            ItemType                       `json:"type"`
	Title           string                         `json:"title"`
	Artist          string                         `json:"artist"`
	ArtID           ArtID                          `json:"art_id"`
	ID              ItemID                         `json:"item_id"`
	ItemType        ItemType                       `json:"item_type"`
	Downloads       map[string]DigitalItemDownload `json:"downloads"`
	DownloadTypeStr ItemType                       `json:"download_type_str"`
}

func (c *Client) Download(d *DigitalItemDownload) (*http.Response, error) {
	req, err := http.NewRequest("GET", d.URL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.c.Do(req)
	if err != nil {
		return resp, err
	}

	if resp.ContentLength == 0 {
		return resp, errors.New("expected file")
	}

	return resp, nil
}

func (c *Client) GetDigitalItem(downloadURL string) (*DigitalItem, error) {
	var blob string
	dd := struct {
		// there is so much garbage here it's insane
		Items []DigitalItem `json:"digital_items"`
	}{}

	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)

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
		return nil, ErrNoData
	}

	err = json.Unmarshal([]byte(blob), &dd)
	if err != nil {
		return nil, err
	}

	return &dd.Items[0], nil
}
