// Package bandcamp provides Web API access to undocumented user API.
package bandcamp

var bc = "https://bandcamp.com"

type (
	FanID       int64
	AlbumID     int64
	ItemID      int64
	BandID      int64
	GenreID     int64
	EncodingsID int64
	PaymentID   int64
	TrackID     int64
	ArtID       int64
	SaleItemID  int64
	ItemArtID   int64
	MerchID     int64
	TralbumID   int64
)

type ItemType string

const (
	// ...Why...???
	AlbumShort ItemType = "a"
	TrackShort ItemType = "t"
	AlbumLong  ItemType = "album"
	TrackLong  ItemType = "track"
)

func (it ItemType) IsAlbum() bool {
	return it == AlbumShort || it == AlbumLong
}

func (it ItemType) IsTrack() bool {
	return it == TrackShort || it == TrackLong
}

func (it ItemType) String() string {
	switch it {
	case AlbumShort, AlbumLong:
		return "Album"
	case TrackShort, TrackLong:
		return "Track"
	default:
		return string(it)
	}
}
