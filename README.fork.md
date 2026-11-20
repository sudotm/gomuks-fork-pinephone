# PinePhone fork

Short record of what this fork is, so the next upstream update and the next release build start from a known commit. Upstream's own `README.md` is unchanged. Layout details are in `PINEPHONE.md`.

## Upstream

| | |
| --- | --- |
| Project | [gomuks/gomuks](https://github.com/gomuks/gomuks) |
| Version | 26.11.1 (`version/version.go`, `BaseVersion`) |
| Tag | [`v0.2611.1`](https://github.com/gomuks/gomuks/releases/tag/v0.2611.1) |
| Commit | `5d21e2c7b31aabbba30efdd33f5882d99972e00e` |

v26.10, the backend this terminal is meant to talk to, is tag `v0.2610.0`. This tree is the newer terminal, `v0.2611.1`.

The local import of that tag is `bf96741` on `main`. PinePhone changes are on `pinephone-layout`.

To move to a newer gomuks tag, fetch that tag and rebase `pinephone-layout` onto it. The fork-only edits are `README.fork.md`, `PINEPHONE.md`, and the files in the layout commits. Leave `README.md` to upstream.

## Shortcuts

Defaults in `tui/config/keybindings.yaml`. Override them in `terminal-keybindings.yaml`. `Ctrl+l` and `Alt+l` stay on the temporary plaintext view.

| Key | Action |
| --- | --- |
| `Alt+1` | `toggle_room_list` |
| `Alt+2` | `toggle_user_list` |
| `Alt+3` | `toggle_timestamps` |
| `Alt+4` | `toggle_compact_messages` |
| `Alt+Up`, `Ctrl+Up` | Previous room |
| `Alt+Down`, `Ctrl+Down` | Next room |
| `Alt+k`, `Ctrl+k` | Room search |
| `Alt+a` | Next room with activity |
| `Alt+l`, `Ctrl+l` | Temporary plaintext view |

## Release build

Published asset: [`gomuks-terminal-v0.2611.1-pinephone.1-linux-aarch64`](https://github.com/sudotm/gomuks-fork-pinephone/releases/tag/v0.2611.1-pinephone.1), built from `pinephone-layout` commit `58f17546543920ac8e84353b067f618b5e55e403` with Go 1.27.4.

`CGO_ENABLED=0` is required. The terminal binary does not use the backend's cgo SQLite or libheif code, and a cgo build would link the host libc instead of running on musl.

```sh
export CGO_ENABLED=0 GOOS=linux GOARCH=arm64
export TAG=v0.2611.1-pinephone.1
export COMMIT=$(git rev-parse HEAD)
export BUILT=$(date -u +%Y-%m-%dT%H:%M:%SZ)
go build -trimpath -ldflags "-s -w -X go.mau.fi/gomuks/version.Tag=$TAG -X go.mau.fi/gomuks/version.Commit=$COMMIT -X go.mau.fi/gomuks/version.BuildTime=$BUILT" \
  -o "dist/gomuks-terminal-$TAG-linux-aarch64" ./cmd/gomuks-terminal
```

Check the result before uploading: ELF 64-bit, machine `0xB7` (aarch64). `BuildTime` changes each run, so a later build will not be byte-identical. `Tag` and `Commit` are what identify it.
