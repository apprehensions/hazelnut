# hazelnut

Simple CLI [bandcamp](https://bandcamp.com) collection synchronizer.

```
$ ./hazelnut -o ~/music  
sjalvmord 7 / 71                                ⠼                                 
a2594097446 475.8 MiB / 518.4 MiB  █████████████████████████████████████████████████▌░░░░░ 2.30MiB/s
a382191878  466.5 MiB / 1.1 GiB    ██████████████████████▌░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░ 2.28MiB/s
a1737495501 266.3 MiB / 498.6 MiB  ████████████████████████████▌░░░░░░░░░░░░░░░░░░░░░░░░░░ 2.24MiB/s
a2783120099 75.5 MiB / 267.6 MiB   ███████████████▌░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░ 2.41MiB/s
t1433347432 29.3 MiB / 102.7 MiB   ███████████████▌░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░ 2.23MiB/s
a3457013254 18.5 MiB / 868.1 MiB   ▌░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░ 2.09MiB/s
```

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
└── home is in your arms
    ├── _ (1433347432).mp3
    └── _ (1987275855)
        ├── 01 _.mp3
        ├── 02 _.mp3
        └── cover.jpg
```
Music belonging to albums or tracks are saved without the artist and album in the filename, since said metadata is already represented in the directory structure.

Albums and tracks with the same name will have their tralbum ID appended to the name.

If the track or album already exists, it is skipped.
