package bandcamp

import (
	"fmt"
	"strconv"
	"time"
)

type FanSummary struct {
	ID         FanID             `json:"fan_id"`
	Collection CollectionSummary `json:"collection_summary"`
}

type CollectionSummary struct {
	Username string `json:"username"`
}

type SaleItemType string

const (
	C SaleItemType = "c"
	P SaleItemType = "p"
	R SaleItemType = "r"
)

type Item struct {
	BandName     string       `json:"band_name"`
	ID           ItemID       `json:"item_id"`
	Title        string       `json:"item_title"`
	Type         ItemType     `json:"item_type"`
	SaleItemID   SaleItemID   `json:"sale_item_id"`
	SaleItemType SaleItemType `json:"sale_item_type"`
}

func (i Item) String() string {
	return i.Type.Short() + strconv.FormatInt(int64(i.ID), 10)
}

func (i Item) Name() string {
	return i.BandName + " - " + i.Title
}

type CollectionItems struct {
	Items []Item `json:"items"`

	// [SaleItemType+SaleItemID]RedownloadURL
	RedownloadURLs map[string]string `json:"redownload_urls"`
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

	r := map[string]interface{}{
		"fan_id":           id,
		"older_than_token": fmt.Sprintf("%d::a::", time.Now().Unix()),
		"count":            2147483647, // get all items
	}

	err := c.Request("POST", "fancollection/1/collection_items", r, &ci)
	if err != nil {
		return nil, err
	}

	return &ci, nil
}
