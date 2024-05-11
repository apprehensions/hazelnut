# hazelnut

Simple Bandcamp collection syncronization & downloader utility.

## Installation
```sh
go install github.com/apprehensions/hazelnut@latest
```

## Usage
`hazelnut` requires the `Cookie` header made on bandcamp web requests.

To retrieve your `Cookie` header, you need to copy it from a [bandcamp.com](https://bandcamp.com/) network request, which can be found in a network request in the 'Request Headers' section under the network requests tab in your browser.

Afterwards, you can copy the copied `Cookie` value to a file named `hazelnut-cookies.txt`, which `hazelnut` uses as the default cookies path (you may choose to change it with the `-cookies` flag).

Example usage:
```
hazelnut -format flac -o Music
```

## Behavior
Music is downloaded to the given output directory (`-o`, default `.`) with this structure:
```
Music
├── Sadness
│   └── atna
│       ├── 01 daydreaming.m4a
│       ├── 02 how bright you shine.m4a
│       ├── 03 hope you never forget.m4a
│       └── cover.jpg
└── ghostslug
    └── Jikininki.flac
```
Music belonging to albums or tracks are saved without the artist and album in the filename, since said metadata is already represented in the directory structure.

If the track's file or album's directory already exists, it is skipped.
