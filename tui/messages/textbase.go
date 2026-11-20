// gomuks - A terminal Matrix client written in Go.
// Copyright (C) 2025 Tulir Asokan
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package messages

import (
	"regexp"

	"github.com/gdamore/tcell/v2"

	"go.mau.fi/gomuks/tui/config"
	"go.mau.fi/gomuks/tui/messages/tstring"
)

// Regular expressions used to split lines when calculating the buffer.
//
// From tview/textview.go
var (
	boundaryPattern     = regexp.MustCompile(`([[:punct:]]\s*|\s+)`)
	bareBoundaryPattern = regexp.MustCompile(`(\s+)`)
	spacePattern        = regexp.MustCompile(`\s+`)
)

func matchBoundaryPattern(bare bool, extract tstring.TString) tstring.TString {
	regex := boundaryPattern
	if bare {
		regex = bareBoundaryPattern
	}
	matches := regex.FindAllStringIndex(extract.String(), -1)
	if len(matches) > 0 {
		if match := matches[len(matches)-1]; len(match) >= 2 {
			if until := match[1]; until < len(extract) {
				extract = extract[:until]
			}
		}
	}
	return extract
}

// CalculateBuffer generates the internal buffer for this message that consists
// of the text of this message split into lines at most as wide as the width
// parameter.
func LayoutPrefStamp(p config.UserPreferences) int {
	stamp := 0
	if p.BareMessageView {
		stamp |= 1
	}
	if p.HideTimestamp {
		stamp |= 2
	}
	if p.DisableImages {
		stamp |= 4
	}
	if p.EnableInlineURLs() {
		stamp |= 8
	}
	return stamp
}

func appendRunes(dst tstring.TString, s string, color tcell.Color) tstring.TString {
	for _, r := range s {
		if color == tcell.ColorDefault {
			dst = append(dst, tstring.NewCell(r))
		} else {
			dst = append(dst, tstring.NewColorCell(r, color))
		}
	}
	return dst
}

// compactPrefix is the author (and optional timestamp) used when the sender
// column is hidden. Service lines keep a timestamp but not the "*" placeholder.
func compactPrefix(prefs config.UserPreferences, msg *UIMessage) tstring.TString {
	if msg == nil || msg.Event == nil {
		return nil
	}
	var prefix tstring.TString
	if !prefs.HideTimestamp {
		if ts := msg.FormatTime(); ts != "" {
			prefix = appendRunes(prefix, ts, msg.TimestampColor())
		}
	}
	if msg.IsService {
		return prefix
	}
	name := msg.GetSenderName()
	if name == "" {
		return prefix
	}
	if len(prefix) > 0 {
		prefix = appendRunes(prefix, " ", tcell.ColorDefault)
	}
	prefix = appendRunes(prefix, "<"+name+">", msg.SenderColor())
	if msg.LastEditRef != nil {
		prefix = appendRunes(prefix, "*", tcell.ColorDarkRed)
	}
	return prefix
}

// compactHeader is a single line naming the author. It is truncated to width
// so a long display name cannot push the message body into another column.
func compactHeader(prefs config.UserPreferences, msg *UIMessage, width int) []tstring.TString {
	if !prefs.BareMessageView || width < 2 {
		return nil
	}
	prefix := compactPrefix(prefs, msg)
	if len(prefix) == 0 {
		return nil
	}
	if prefix.RuneWidth() > width {
		prefix = prefix.Truncate(width)
	}
	if len(prefix) == 0 {
		return nil
	}
	return []tstring.TString{prefix}
}

func calculateBufferWithText(prefs config.UserPreferences, text tstring.TString, width int, msg *UIMessage) []tstring.TString {
	if width < 2 {
		return nil
	}

	var buffer []tstring.TString

	if prefs.BareMessageView {
		if prefix := compactPrefix(prefs, msg); len(prefix) > 0 {
			if len(text) > 0 {
				text = prefix.Append(" ").AppendTString(text)
			} else {
				text = prefix
			}
		}
	}

	forcedLinebreaks := text.Split('\n')
	newlines := 0
	for _, str := range forcedLinebreaks {
		if len(str) == 0 && newlines < 1 {
			buffer = append(buffer, tstring.TString{})
			newlines++
		} else {
			newlines = 0
		}
		// Adapted from tview/textview.go#reindexBuffer()
		for len(str) > 0 {
			extract := str.Truncate(width)
			if len(extract) < len(str) {
				if spaces := spacePattern.FindStringIndex(str[len(extract):].String()); spaces != nil && spaces[0] == 0 {
					// Keep the break space on this line only when it still fits.
					// Absorbing it unconditionally made the line wider than width.
					candidate := str[:len(extract)+spaces[1]]
					if candidate.RuneWidth() <= width {
						extract = candidate
					}
				}
				extract = matchBoundaryPattern(prefs.BareMessageView, extract)
			}
			if len(extract) == 0 {
				extract = str[:1]
			}
			buffer = append(buffer, extract)
			str = str[len(extract):]
		}
	}
	return buffer
}
