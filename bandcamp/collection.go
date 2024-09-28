package bandcamp

import (
	"fmt"
	"math"
	"time"
)

type FanSummary struct {
	ID         FanID             `json:"fan_id"`
	Collection CollectionSummary `json:"collection_summary"`
}

type CollectionSummary struct {
	ID       FanID                        `json:"fan_id"`
	Username string                       `json:"username"`
	URL      string                       `json:"url"`
	Lookup   map[string]CollectionSummary `json:"tralbum_lookup"`
	Follows  FollowsFans                  `json:"follows"`
}

type FollowsFans struct {
	Following map[FanID]bool `json:"following"`
}

type CollectionSummaryItem struct {
	Type   ItemType `json:"item_type"`
	ID     ItemID   `json:"item_id"`
	BandID BandID   `json:"band_id"`
}

// [SaleItemType+SaleItemID]URL
type DownloadURLs map[string]string

type ItemArt struct {
	ArtID    ArtID  `json:"art_id"`
	ThumbURL string `json:"thumb_url"`
	URL      string `json:"url"`
}

type SaleItemType string

type URLHints struct {
	// Unknown fields ommitted
	ItemType  ItemType `json:"item_type"`
	Slug      string   `json:"slug"`
	Subdomain string   `json:"subdomain"`
}

type CollectionItem struct {
	// Many fields ommitted - out of scope
	AlsoCollectedCount int64        `json:"also_collected_count"`
	BandID             BandID       `json:"band_id"`
	BandName           string       `json:"band_name"`
	BandURL            string       `json:"band_url"`
	FanID              FanID        `json:"fan_id"`
	HasDigitalDownload bool         `json:"has_digital_download"`
	Hidden             bool         `json:"hidden"`
	Art                ItemArt      `json:"item_art"`   // Replacement for "item_art_"*
	ID                 ItemID       `json:"item_id"`    // Replacement for "album_id" & "tralbum_id"
	Title              string       `json:"item_title"` // Replacement for "album_title"
	Type               ItemType     `json:"item_type"`  // Replacement(?) for "tralbum_type"
	URL                string       `json:"item_url"`
	SaleItemID         SaleItemID   `json:"sale_item_id"`
	SaleItemType       SaleItemType `json:"sale_item_type"`
	Token              string       `json:"token"`
}

type Item struct {
	DownloadURL string
	CollectionItem
}

type CollectionItems struct {
	LastToken      string           `json:"last_token"`
	Items          []CollectionItem `json:"items"`
	MoreAvailable  bool             `json:"more_available"`
	RedownloadURLs DownloadURLs     `json:"redownload_urls"`
}

func (ci CollectionItem) String() string {
	return fmt.Sprintf("%s %s%d", ci.BandName, ci.Type.Short(), ci.ID)
}

func (c *Client) GetCollectionSummary() (*FanSummary, error) {
	var fs FanSummary

	err := c.Request("GET", "fan/2/collection_summary", nil, &fs)
	if err != nil {
		return nil, err
	}

	return &fs, nil
}

func (c *Client) GetCollectionItems(id FanID) (*CollectionItems, error) {
	var ci CollectionItems

	req := map[string]interface{}{
		"fan_id":           id,
		"older_than_token": fmt.Sprintf("%d::a::", time.Now().Unix()),
		"count":            math.MaxInt32,
	}

	err := c.Request("POST", "fancollection/1/collection_items", req, &ci)
	if err != nil {
		return nil, err
	}

	return &ci, nil
}

func (c *Client) GetItems(id FanID) ([]Item, error) {
	ci, err := c.GetCollectionItems(id)
	if err != nil {
		return nil, err
	}

	var items []Item

	for si, u := range ci.RedownloadURLs {
		for _, i := range ci.Items {
			if si != fmt.Sprint(i.SaleItemType, i.SaleItemID) {
				continue
			}

			items = append(items, Item{
				CollectionItem: i,
				DownloadURL:    u,
			})
		}
	}

	return items, nil
}
