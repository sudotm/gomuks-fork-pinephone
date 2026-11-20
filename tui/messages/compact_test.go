// gomuks - A terminal Matrix client written in Go.
// Copyright (C) 2025 Tulir Asokan
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package messages

import (
	"strings"
	"testing"
	"time"

	"go.mau.fi/util/jsontime"

	"go.mau.fi/gomuks/pkg/hicli/database"
	"go.mau.fi/gomuks/tui/config"
	"go.mau.fi/gomuks/tui/messages/html"
	"go.mau.fi/gomuks/tui/messages/tstring"
)

func testMessage(name string, text tstring.TString) *UIMessage {
	return &UIMessage{
		Event: &database.Event{
			Timestamp: jsontime.UM(time.Date(2026, 1, 2, 15, 4, 5, 0, time.UTC)),
			Sender:    "@alice:example.invalid",
		},
		OverrideSenderName: name,
		Renderer: &ExpandedTextMessage{
			Text: text,
		},
	}
}

func bufferText(msg *UIMessage) string {
	lines := msg.Renderer.(*ExpandedTextMessage).buffer
	parts := make([]string, len(lines))
	for i, line := range lines {
		parts[i] = line.String()
	}
	return strings.Join(parts, "\n")
}

func TestCompactTextReleasesSenderColumn(t *testing.T) {
	body := tstring.NewTString("hello from the pinephone")
	column := testMessage("Alice", body)
	column.CalculateBuffer(config.UserPreferences{}, 40)
	if strings.Contains(bufferText(column), "<Alice>") {
		t.Fatalf("column layout folded the sender into the body: %q", bufferText(column))
	}

	compact := testMessage("Alice", body)
	compact.CalculateBuffer(config.UserPreferences{BareMessageView: true, HideTimestamp: true}, 40)
	got := bufferText(compact)
	if !strings.Contains(got, "<Alice>") || !strings.Contains(got, "hello from the pinephone") {
		t.Fatalf("compact body = %q", got)
	}
	if strings.Contains(got, compact.FormatTime()) {
		t.Fatalf("hidden timestamp still rendered: %q", got)
	}

	withTime := testMessage("Alice", body)
	withTime.CalculateBuffer(config.UserPreferences{BareMessageView: true}, 40)
	if !strings.Contains(bufferText(withTime), withTime.FormatTime()) || !strings.Contains(bufferText(withTime), "<Alice>") {
		t.Fatalf("timestamped compact body = %q", bufferText(withTime))
	}

	// Same width, different timestamp preference, must not reuse the old wrap.
	withTime.CalculateBuffer(config.UserPreferences{BareMessageView: true, HideTimestamp: true}, 40)
	if strings.Contains(bufferText(withTime), withTime.FormatTime()) {
		t.Fatalf("timestamp toggle reused the previous buffer: %q", bufferText(withTime))
	}
}

func TestCompactWrapsLongURLParagraphsAndUnicode(t *testing.T) {
	url := "https://example.invalid/this/is/a/very/long/unbroken/path"
	paragraphs := "first paragraph\n\nsecond paragraph " + url
	msg := testMessage("Alice With A Long Display Name", tstring.NewTString(paragraphs))
	const width = 24
	msg.CalculateBuffer(config.UserPreferences{BareMessageView: true, HideTimestamp: true}, width)
	lines := msg.Renderer.(*ExpandedTextMessage).buffer
	if len(lines) < 3 {
		t.Fatalf("expected wrapped paragraphs, got %d lines: %#v", len(lines), bufferText(msg))
	}
	blank := false
	for _, line := range lines {
		if line.RuneWidth() > width {
			t.Fatalf("line wider than %d: %q (%d)", width, line.String(), line.RuneWidth())
		}
		if len(line) == 0 {
			blank = true
		}
	}
	if !blank {
		t.Fatal("blank line between paragraphs was dropped")
	}
	flat := strings.ReplaceAll(bufferText(msg), "\n", "")
	if !strings.Contains(flat, "<Alice With A Long Display Name>") {
		t.Fatalf("long name missing from compact body: %q", bufferText(msg))
	}

	var wide tstring.TString
	for i := 0; i < 4; i++ {
		wide = append(wide, tstring.NewCell('😀'))
	}
	for i := 0; i < 8; i++ {
		wide = append(wide, tstring.NewCell('e'), tstring.NewCell('\u0301'))
	}
	unicodeMsg := testMessage("A😀", wide)
	unicodeMsg.CalculateBuffer(config.UserPreferences{BareMessageView: true, HideTimestamp: true}, 5)
	for _, line := range unicodeMsg.Renderer.(*ExpandedTextMessage).buffer {
		if line.RuneWidth() > 5 {
			t.Fatalf("unicode line wider than 5: %q (%d)", line.String(), line.RuneWidth())
		}
	}
}

func TestServiceAndRedactedCompactHeaders(t *testing.T) {
	service := testMessage("*", tstring.NewTString("Date changed to January 2, 2026"))
	service.IsService = true
	service.CalculateBuffer(config.UserPreferences{BareMessageView: true}, 60)
	if strings.Contains(bufferText(service), "<*>") {
		t.Fatalf("service line included a placeholder sender: %q", bufferText(service))
	}

	redacted := &UIMessage{
		Event: &database.Event{
			Timestamp: jsontime.UM(time.Date(2026, 1, 2, 15, 4, 5, 0, time.UTC)),
		},
		OverrideSenderName: "Alice",
		Renderer:           &RedactedMessage{},
	}
	redacted.CalculateBuffer(config.UserPreferences{BareMessageView: true, HideTimestamp: true}, 40)
	if redacted.Height() != 2 {
		t.Fatalf("redacted compact height = %d, want header plus bar", redacted.Height())
	}
	header := redacted.Renderer.(*RedactedMessage).header
	if len(header) != 1 || !strings.Contains(header[0].String(), "<Alice>") {
		t.Fatalf("redacted header = %#v", header)
	}
}

func TestHTMLCompactHeaderUsesFullBodyWidth(t *testing.T) {
	body := strings.Repeat("a", 40)
	root := html.NewTextEntity(body)
	msg := &UIMessage{
		Event: &database.Event{
			Timestamp: jsontime.UM(time.Date(2026, 1, 2, 15, 4, 5, 0, time.UTC)),
		},
		OverrideSenderName: strings.Repeat("Name", 20),
		Renderer: &HTMLMessage{
			Root: root,
		},
	}
	const width = 20
	msg.CalculateBuffer(config.UserPreferences{BareMessageView: true, HideTimestamp: true}, width)
	hw := msg.Renderer.(*HTMLMessage)
	if len(hw.header) != 1 {
		t.Fatalf("header lines = %d, want 1", len(hw.header))
	}
	if hw.header[0].RuneWidth() > width {
		t.Fatalf("header width = %d", hw.header[0].RuneWidth())
	}
	if !strings.Contains(hw.header[0].String(), "<Name") {
		t.Fatalf("header = %q", hw.header[0].String())
	}
	if root.Height() != 2 {
		t.Fatalf("body height = %d, want 2 lines at width %d", root.Height(), width)
	}
	if hw.Height() != 3 {
		t.Fatalf("message height = %d, want header plus two body lines", hw.Height())
	}
	if strings.Contains(root.PlainText(), "<Name") {
		t.Fatalf("header was inserted into the body: %q", root.PlainText())
	}
}
