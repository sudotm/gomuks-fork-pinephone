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
	"github.com/gdamore/tcell/v2"
	"go.mau.fi/mauview"
	"maunium.net/go/mautrix/event"

	"go.mau.fi/gomuks/pkg/hicli/database"
	"go.mau.fi/gomuks/pkg/rpc/store"
	"go.mau.fi/gomuks/tui/config"
	"go.mau.fi/gomuks/tui/messages/html"
	"go.mau.fi/gomuks/tui/messages/tstring"
)

type HTMLMessage struct {
	Root      html.Entity
	TextColor tcell.Color
	header    []tstring.TString
}

func NewHTMLMessage(room *store.RoomStore, evt *database.Event, content *event.MessageEventContent, root html.Entity) *UIMessage {
	return newUIMessage(room, evt, content, "", &HTMLMessage{
		Root: root,
	})
}

func (hw *HTMLMessage) Clone() MessageRenderer {
	return &HTMLMessage{
		Root: hw.Root.Clone(),
	}
}

func (hw *HTMLMessage) Draw(screen mauview.Screen, msg *UIMessage) {
	if hw.TextColor != tcell.ColorDefault {
		hw.Root.AdjustStyle(func(style tcell.Style) tcell.Style {
			fg, _, _ := style.Decompose()
			if fg == tcell.ColorDefault {
				return style.Foreground(hw.TextColor)
			}
			return style
		}, html.AdjustStyleReasonNormal)
	}
	screen.Clear()
	y := 0
	for _, line := range hw.header {
		line.Draw(screen, 0, y)
		y++
	}
	body := mauview.Screen(screen)
	if y > 0 {
		w, h := screen.Size()
		if h < y {
			h = y
		}
		body = mauview.NewProxyScreen(screen, 0, y, w, h-y)
	}
	hw.Root.Draw(body, html.DrawContext{IsSelected: msg.IsSelected})
}

func (hw *HTMLMessage) OnKeyEvent(event mauview.KeyEvent) bool {
	return false
}

func (hw *HTMLMessage) OnMouseEvent(event mauview.MouseEvent) bool {
	return false
}

func (hw *HTMLMessage) OnPasteEvent(event mauview.PasteEvent) bool {
	return false
}

func (hw *HTMLMessage) CalculateBuffer(preferences config.UserPreferences, width int, msg *UIMessage) {
	if width < 2 {
		return
	}
	// HTML from Matrix is usually a block element, so the sender stays on a
	// one-line header instead of a permanent column. The body then uses width.
	hw.header = compactHeader(preferences, msg, width)
	startX := 0
	hw.TextColor = msg.TextColor()
	hw.Root.CalculateBuffer(width, startX, html.DrawContext{
		IsSelected:   msg.IsSelected,
		BareMessages: preferences.BareMessageView,
	})
}

func (hw *HTMLMessage) Height() int {
	return len(hw.header) + hw.Root.Height()
}

func (hw *HTMLMessage) PlainText() string {
	return hw.Root.PlainText()
}

func (hw *HTMLMessage) NotificationContent() string {
	return hw.Root.PlainText()
}

func (hw *HTMLMessage) String() string {
	return hw.Root.String()
}
