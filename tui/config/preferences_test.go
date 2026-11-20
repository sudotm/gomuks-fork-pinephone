// gomuks - A terminal Matrix client written in Go.
// Copyright (C) 2020 Tulir Asokan
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"codeberg.org/tslocum/cbind"
	"gopkg.in/yaml.v3"
)

func TestLoadMissingFileUsesDefaults(t *testing.T) {
	cfg := NewConfig()
	cfg.Dir = t.TempDir()
	cfg.Load()
	if cfg.Preferences != (UserPreferences{}) {
		t.Fatalf("missing file preferences = %+v, want zero values", cfg.Preferences)
	}
	if cfg.Server != "" || cfg.Username != "" || cfg.Password != "" {
		t.Fatalf("missing file filled connection fields: %+v", cfg)
	}
}

func TestLoadPartialPreferences(t *testing.T) {
	dir := t.TempDir()
	original := "" +
		"# connection comment\n" +
		"server: http://127.0.0.1:29325\n" +
		"username: demo\n" +
		"password: dummy-password\n" +
		"custom_client_note: keep-this\n" +
		"preferences:\n" +
		"  hide_room_list: true\n"
	path := filepath.Join(dir, "terminal.yaml")
	if err := os.WriteFile(path, []byte(original), 0600); err != nil {
		t.Fatal(err)
	}

	cfg := NewConfig()
	cfg.Dir = dir
	cfg.Load()
	if cfg.Server != "http://127.0.0.1:29325" || cfg.Username != "demo" || cfg.Password != "dummy-password" {
		t.Fatalf("connection settings = %s %s %s", cfg.Server, cfg.Username, cfg.Password)
	}
	if !cfg.Preferences.HideRoomList {
		t.Fatal("hide_room_list was not loaded")
	}
	if cfg.Preferences.HideUserList || cfg.Preferences.HideTimestamp || cfg.Preferences.BareMessageView {
		t.Fatalf("unset preferences were not left false: %+v", cfg.Preferences)
	}

	cfg.Preferences.HideUserList = true
	if err := cfg.SavePreferences(); err != nil {
		t.Fatal(err)
	}
	saved, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(saved)
	for _, want := range []string{
		"server: http://127.0.0.1:29325",
		"username: demo",
		"password: dummy-password",
		"custom_client_note: keep-this",
		"hide_room_list: true",
		"hide_user_list: true",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("saved config missing %q:\n%s", want, text)
		}
	}
	if !strings.Contains(text, "connection comment") {
		t.Fatalf("saved config dropped the existing comment:\n%s", text)
	}

	reloaded := NewConfig()
	reloaded.Dir = dir
	reloaded.Load()
	if !reloaded.Preferences.HideRoomList || !reloaded.Preferences.HideUserList {
		t.Fatalf("reloaded preferences = %+v", reloaded.Preferences)
	}
	if reloaded.Preferences.HideTimestamp || reloaded.Preferences.BareMessageView {
		t.Fatalf("reloaded unset preferences = %+v", reloaded.Preferences)
	}
	if reloaded.Password != "dummy-password" || reloaded.Server != cfg.Server {
		t.Fatalf("reloaded connection settings changed: %+v", reloaded)
	}
}

func TestSavePreferencesLeavesMalformedFile(t *testing.T) {
	dir := t.TempDir()
	original := "server: [\n"
	path := filepath.Join(dir, "terminal.yaml")
	if err := os.WriteFile(path, []byte(original), 0600); err != nil {
		t.Fatal(err)
	}
	cfg := NewConfig()
	cfg.Dir = dir
	cfg.Preferences.HideRoomList = true
	err := cfg.SavePreferences()
	if err == nil {
		t.Fatal("expected malformed terminal.yaml to fail")
	}
	if !strings.Contains(err.Error(), "malformed") {
		t.Fatalf("error = %v, want a malformed-file error", err)
	}
	saved, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(saved) != original {
		t.Fatalf("malformed file was modified:\n%s", saved)
	}
}

func TestLoadMalformedPanics(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "terminal.yaml")
	if err := os.WriteFile(path, []byte("server: [\n"), 0600); err != nil {
		t.Fatal(err)
	}
	cfg := NewConfig()
	cfg.Dir = dir
	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatal("expected load to panic")
		}
		msg := recovered.(error).Error()
		if !strings.Contains(msg, "terminal.yaml") {
			t.Fatalf("panic = %s, want the config path", msg)
		}
		saved, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if string(saved) != "server: [\n" {
			t.Fatalf("load rewrote the broken file:\n%s", saved)
		}
	}()
	cfg.Load()
}

func TestDefaultLayoutBindings(t *testing.T) {
	var raw RawKeybindings
	if err := yaml.Unmarshal([]byte(DefaultKeybindings), &raw); err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"Alt+1":     "toggle_room_list",
		"Alt+2":     "toggle_user_list",
		"Alt+3":     "toggle_timestamps",
		"Alt+4":     "toggle_compact_messages",
		"Ctrl+l":    "show_bare",
		"Alt+l":     "show_bare",
		"Ctrl+Down": "next_room",
		"Alt+Down":  "next_room",
		"Ctrl+Up":   "prev_room",
		"Alt+Up":    "prev_room",
		"Ctrl+k":    "search_rooms",
		"Alt+k":     "search_rooms",
		"Alt+a":     "next_active_room",
	}
	for shortcut, action := range want {
		if raw.Main[shortcut] != action {
			t.Fatalf("%s = %q, want %q", shortcut, raw.Main[shortcut], action)
		}
	}
	actions := map[string]string{}
	for shortcut, action := range raw.Main {
		if prev, ok := actions[action]; ok && (action == "toggle_room_list" || action == "toggle_user_list" || action == "toggle_timestamps" || action == "toggle_compact_messages") {
			t.Fatalf("%s bound to both %s and %s", action, prev, shortcut)
		}
		actions[action] = shortcut
		if _, _, _, err := cbind.Decode(shortcut); err != nil {
			t.Fatalf("decode %s: %v", shortcut, err)
		}
	}
}
