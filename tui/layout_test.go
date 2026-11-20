// gomuks - A terminal Matrix client written in Go.
// Copyright (C) 2025 Tulir Asokan
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package tui

import (
	"testing"

	"go.mau.fi/gomuks/tui/config"
	"go.mau.fi/gomuks/tui/messages"
)

func conversationBodyWidth(termWidth int, prefs config.UserPreferences) int {
	roomWidth := termWidth
	if !prefs.HideRoomList {
		roomWidth -= RoomListColumnWidth + RoomListBorderWidth
	}
	if roomWidth < 0 {
		roomWidth = 0
	}
	content := roomContentWidth(roomWidth, prefs.HideUserList)
	return messageTextWidth(content, prefs, 15, len(messages.TimeFormat))
}

func TestConversationBodyWidth(t *testing.T) {
	pinephone := config.UserPreferences{
		HideRoomList:    true,
		HideUserList:    true,
		HideTimestamp:   true,
		BareMessageView: true,
	}
	cases := []struct {
		name  string
		cols  int
		prefs config.UserPreferences
		want  int
	}{
		{name: "80 default columns", cols: 80, want: 6},
		{name: "80 sidebars hidden", cols: 80, prefs: config.UserPreferences{HideRoomList: true, HideUserList: true}, want: 53},
		{name: "80 sidebars and timestamps hidden", cols: 80, prefs: config.UserPreferences{HideRoomList: true, HideUserList: true, HideTimestamp: true}, want: 62},
		{name: "80 pinephone compact", cols: 80, prefs: pinephone, want: 80},
		{name: "60 pinephone compact", cols: 60, prefs: pinephone, want: 60},
		{name: "120 pinephone compact", cols: 120, prefs: pinephone, want: 120},
		{name: "60 default columns", cols: 60, want: 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := conversationBodyWidth(tc.cols, tc.prefs); got != tc.want {
				t.Fatalf("body width = %d, want %d", got, tc.want)
			}
		})
	}
	if got := roomContentWidth(80, true); got != 80 {
		t.Fatalf("hidden member list width = %d, want 80", got)
	}
	if got := roomContentWidth(80, false); got != 80-StaticHorizontalSpace {
		t.Fatalf("shown member list width = %d, want %d", got, 80-StaticHorizontalSpace)
	}
}
