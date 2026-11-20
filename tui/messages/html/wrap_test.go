// gomuks - A terminal Matrix client written in Go.
// Copyright (C) 2020 Tulir Asokan
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package html

import (
	"strings"
	"testing"

	"github.com/mattn/go-runewidth"
)

func TestTextEntityWrapsURLEmojiAndCombiningMarks(t *testing.T) {
	url := "see https://example.invalid/this/is/a/very/long/unbroken/path now"
	entity := NewTextEntity(url + " 😀😀😀 " + strings.Repeat("e\u0301", 6))
	const width = 16
	entity.CalculateBuffer(width, 0, DrawContext{BareMessages: true})
	if entity.Height() < 2 {
		t.Fatalf("height = %d, want the long URL to wrap", entity.Height())
	}
	for _, line := range entity.buffer {
		if got := runewidth.StringWidth(line); got > width {
			t.Fatalf("line width %d > %d: %q", got, width, line)
		}
	}
}

func TestSpoilerStillCalculatesHiddenAndVisibleText(t *testing.T) {
	visible := &ContainerEntity{
		BaseEntity: &BaseEntity{Tag: "span"},
		Children:   []Entity{NewTextEntity("secret text")},
	}
	spoiler := NewSpoilerEntity(visible, "cw")
	spoiler.CalculateBuffer(30, 0, DrawContext{BareMessages: true})
	if spoiler.Height() < 1 {
		t.Fatal("spoiler height was empty")
	}
	if spoiler.PlainText() != "spoiler: cw" {
		t.Fatalf("spoiler label = %q", spoiler.PlainText())
	}
	if !strings.Contains(spoiler.visible.PlainText(), "secret text") {
		t.Fatalf("spoiler contents = %q", spoiler.visible.PlainText())
	}
}
