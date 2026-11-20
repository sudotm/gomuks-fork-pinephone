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
	"go.mau.fi/gomuks/tui/messages/tstring"
)

type RedactedMessage struct {
	header []tstring.TString
}

func NewRedactedMessage(evt *database.Event, room *store.RoomStore) *UIMessage {
	return newUIMessage(room, evt, &event.MessageEventContent{}, "", &RedactedMessage{})
}

func (msg *RedactedMessage) Clone() MessageRenderer {
	return &RedactedMessage{}
}

func (msg *RedactedMessage) NotificationContent() string {
	return ""
}

func (msg *RedactedMessage) PlainText() string {
	return "[redacted]"
}

func (msg *RedactedMessage) String() string {
	return "&messages.RedactedMessage{}"
}

func (msg *RedactedMessage) CalculateBuffer(prefs config.UserPreferences, width int, uiMsg *UIMessage) {
	msg.header = compactHeader(prefs, uiMsg, width)
}

func (msg *RedactedMessage) Height() int {
	if len(msg.header) == 0 {
		return 1
	}
	return len(msg.header) + 1
}

const RedactionChar = '█'
const RedactionMaxWidth = 40

var RedactionStyle = tcell.StyleDefault.Foreground(tcell.NewRGBColor(50, 0, 0))

func (msg *RedactedMessage) Draw(screen mauview.Screen, _ *UIMessage) {
	y0 := 0
	for _, line := range msg.header {
		line.Draw(screen, 0, y0)
		y0++
	}
	w, _ := screen.Size()
	for x := 0; x < w && x < RedactionMaxWidth; x++ {
		screen.SetContent(x, y0, RedactionChar, nil, RedactionStyle)
	}
}
