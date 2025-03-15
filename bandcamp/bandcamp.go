// Package bandcamp provides stripped required Web API access to undocumented user API.
package bandcamp

import (
	"fmt"
	"time"
)

var bc = "https://bandcamp.com"

type (
	FanID      int64
	ItemID     int64
	SaleItemID int64
)

type ItemType int

const (
	Album ItemType = iota
	Track
)

type ItemRelease struct {
	time.Time
}

func (t ItemType) String() string {
	switch t {
	case Album:
		return "album"
	case Track:
		return "track"
	default:
		panic("unknown ItemType")
	}
}

func (t ItemType) Short() string {
	switch t {
	case Album:
		return "a"
	case Track:
		return "t"
	default:
		panic("unknown ItemType")
	}
}

func (t ItemType) MarshalJSON() ([]byte, error) {
	return []byte(t.String()), nil
}

func (t *ItemType) UnmarshalJSON(b []byte) error {
	switch string(b[1 : len(b)-1]) {
	case "album", "a":
		*t = Album
	case "track", "t":
		*t = Track
	default:
		return fmt.Errorf("invalid ItemType: %s", string(b))
	}
	return nil
}

func (r ItemRelease) MarshalJSON() ([]byte, error) {
	return []byte(r.Format("02 Jan 2006 15:04:05 GMT")), nil
}

func (r *ItemRelease) UnmarshalJSON(b []byte) error {
	date, err := time.Parse(`"02 Jan 2006 15:04:05 GMT"`, string(b))
	if err != nil {
		return err
	}
	*r = ItemRelease{Time: date}
	return nil
}
