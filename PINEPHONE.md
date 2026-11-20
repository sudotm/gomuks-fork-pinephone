# PinePhone layout for the gomuks terminal

This is a small fork of the gomuks **v26.11.1** terminal frontend. It keeps the upstream AGPL-3.0-or-later license and Tulir Asokan's copyright. The layout changes are under the same license. The backend, encryption, and message-send behavior are unchanged.

The release tags are calendar versions with a `v0.` prefix: v26.10 is `v0.2610.0` and v26.11.1 is `v0.2611.1`. This tree matches `v0.2611.1` (`version/version.go` reports `26.11.1`). v26.11 and v26.11.1 do not list terminal-UI changes. The terminal talks to the backend over the existing JSON-RPC websocket and does not call the v26.11-only backend methods (room directory, room upgrade, FCM, OAuth step-up). Unknown JSON fields are ignored. That is compatible with a v26.10 backend in principle. This session did not connect a built binary to a live v26.10 server; keep the previous `gomuks-terminal` binary so you can switch back.

## What was wrong

At 80 columns the default layout is:

- room list: 25 columns, plus a 1-column border
- member list: 20 columns, plus a 1-column border
- timestamp: 8 columns (`15:04:05`), plus a gap
- sender name: 15 columns, plus a gap

The message body at 80×19 is about **6 columns**. Hiding both sidebars, timestamps, and the sender column gives the body all **80** columns.

`HideRoomList`, `HideUserList`, `HideTimestamp`, and `BareMessageView` already existed on `UserPreferences` and were read while drawing. They were tagged `yaml:"-"`, so `terminal.yaml` never loaded or saved them, and no keybinding changed them. `BareMessageView` also still painted the 15-column sender name on top of the message.

## Configuration

Preferences live in the existing `terminal.yaml`, under `preferences`. There is no second config file. A file with no `preferences` key still loads. Omitted booleans default to `false`, which is the ordinary upstream layout.

```yaml
server: http://127.0.0.1:29325
username: demo
password: dummy-password
preferences:
    hide_room_list: true
    hide_user_list: true
    hide_timestamp: true
    bare_message_view: true
```

| Field | Default | Effect |
| --- | --- | --- |
| `hide_room_list` | `false` | Hide the left room list and its border. The topic line shows the room name. |
| `hide_user_list` | `false` | Hide the right member list and its border, and release those columns. |
| `hide_timestamp` | `false` | Hide the timestamp column. In compact mode, also hide the inline time. |
| `bare_message_view` | `false` | Compact messages: no sender column and no scrollbar gutter. |
| `inline_url_mode` | empty | `enable`, `disable`, or empty for the existing terminal detection. |

Other `preferences` booleans (`disable_images`, and the rest of `UserPreferences`) are unchanged and also default to `false`.

Toggling a layout setting rewrites only the `preferences` key. Connection fields and unknown keys are kept. The new file is written to a temporary file and renamed into place, so a marshal or parse failure leaves the previous `terminal.yaml` as it was. A malformed `terminal.yaml` still stops startup, with the path in the error, and is not overwritten.

`Ctrl+l` and `Alt+l` still open the temporary plaintext view (`show_bare`). That is separate from `bare_message_view`.

## Shortcuts

Default bindings are in `tui/config/keybindings.yaml`. Override them in `terminal-keybindings.yaml` in the gomuks config directory. These do not collide with `show_bare`.

| Key | Action | Effect |
| --- | --- | --- |
| `Alt+1` | `toggle_room_list` | Show or hide the room list |
| `Alt+2` | `toggle_user_list` | Show or hide the member list |
| `Alt+3` | `toggle_timestamps` | Show or hide timestamps |
| `Alt+4` | `toggle_compact_messages` | Switch the sender column and compact messages |

Room navigation from this release, including while the room list is hidden:

| Key | Action |
| --- | --- |
| `Ctrl+Down`, `Alt+Down` | Next room |
| `Ctrl+Up`, `Alt+Up` | Previous room |
| `Ctrl+k`, `Alt+k` | Room search |
| `Alt+a` | Next room with activity |
| `PageUp`, `PageDown`, `Ctrl+p`, `Ctrl+n` | Scroll the timeline. Scrolling to the top loads older messages |
| `Enter` | Send |
| `Ctrl+Enter`, `Alt+Enter` | New line in the composer |
| `Escape` | Clear reply, edit, and selection |
| `Ctrl+l`, `Alt+l` | Temporary plaintext view |

A toggle updates the preference, asks the open room to reflow, redraws, and saves. It does not reconnect, restart, move focus, or clear the composer. The topic line is the room name while the room list is hidden.

## Compact messages

`BareMessageView` is reused. Two gaps in that path are filled:

- The sender column, timestamp column, edit marker, and scrollbar gutter are not drawn in compact mode, so they no longer cover the body or keep a blank gutter.
- Plain-text buffers put `timestamp <sender>` on the first line and wrap the rest at the full width. Formatted messages (the normal Matrix path) and redacted messages get one header line, truncated to the available width, and the body uses the full width underneath. Spoilers still use the existing spoiler entity.

Long display names and URLs wrap or truncate inside that width. They do not keep a fixed column.

The topic and status rows are still one each, and the composer still grows up to five rows. That leaves room to type a multi-line message at 19 rows. There is no extra shortcut legend on screen.

## Build

The terminal command is `./cmd/gomuks-terminal`. It does not import the backend's cgo SQLite driver or libheif. `CGO_ENABLED=0` is the right PinePhone build: the binary is statically linked and runs on musl (Alpine / postmarketOS) without a cross C toolchain. `GOARCH=arm64` alone is not what makes it musl-safe; leaving cgo on would link the host libc.

From `gomuks-0.2611.1`, with Go 1.26 or newer (the module asks for the `go1.27.4` toolchain):

```sh
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -o gomuks-terminal-pinephone-linux-arm64 ./cmd/gomuks-terminal
```

Check the ELF header before copying it to the phone. `aarch64` is machine type `0xB7`:

```sh
# Linux
readelf -h gomuks-terminal-pinephone-linux-arm64 | head
```

On the PinePhone, install it beside the original binary:

```sh
install -m 0755 gomuks-terminal-pinephone-linux-arm64 "$HOME/bin/gomuks-terminal-pinephone"
```

Run it with the same config directory the current frontend already uses. Do not rename or delete the original binary.

```sh
GOMUKS_CONFIG_HOME="$HOME/.config/gomuks" "$HOME/bin/gomuks-terminal-pinephone"
```

Rollback is the original command, for example `gomuks-terminal`.

Built here with Go 1.27.4:

```text
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -o dist/gomuks-terminal-pinephone-linux-arm64 ./cmd/gomuks-terminal
```

The result is `dist/gomuks-terminal-pinephone-linux-arm64` (19,953,027 bytes). Its header is ELF 64-bit, machine `0xB7` (aarch64). `go version -m` reports `go1.27.4`. A Windows amd64 development build, `dist/gomuks-terminal-windows-amd64.exe`, also compiled. Neither binary was started against a backend, and nothing was copied to the server or the phone. The binaries are build output and are not part of the commits.

## Tests performed

Automated with Go 1.27.4 (`CGO_ENABLED=0`). `go test ./tui/config ./tui/messages/html ./tui/messages ./tui -count=1` passed:

- Missing `terminal.yaml` loads empty preferences.
- A partial `preferences` map loads, and a later save keeps connection fields, an unknown key, and a comment.
- Malformed YAML is reported and not overwritten.
- Default keybindings parse, including the four toggles, and `Ctrl+l` / `Alt+l` stay on `show_bare`.
- Body width at 80, 60, and 120 columns for the default and compact layouts.
- Compact wrapping for a long name, a long URL, paragraphs, emoji, and a combining character.
- HTML compact mode keeps a one-line header and wraps the body at the full width.
- Spoiler calculation still returns the content warning and the hidden text.

Not run here:

- A live 80×19, 60×19, or 120×30 session in foot, including resize while composing.
- Scrolling, backfill, replies, reactions, and edits against a v26.10 backend.
- The temporary plaintext view on device.

The legacy `/toggle` and `/help` commands are not part of this frontend. `NewHelpModal` is never opened. Use the keys above.

## Manual check

1. Start from an existing `terminal.yaml` that has no `preferences` key. The layout should match the original frontend.
2. Press `Alt+1`, `Alt+2`, `Alt+3`, and `Alt+4` one at a time, then all four. The composer text should stay. The sidebars should not leave an empty column.
3. Quit and start again. The same four flags should come back.
4. Resize the terminal to 80×19, 60×19, and 120×30 while a long message is on screen and while the composer has several lines.
5. With the room list hidden, use `Alt+k` to search and `Alt+Up` / `Alt+Down` to change rooms. The topic line should show the room name.
6. Open a thread with a long name, a long URL, several paragraphs, emoji, a combining character, a code block, a spoiler, a reply, and a message taller than the screen.
7. Press `Alt+l` and confirm the plaintext view still returns on Enter.
