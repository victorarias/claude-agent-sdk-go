// Copyright (C) 2025 Claude Agent SDK Go Contributors
// SPDX-License-Identifier: GPL-3.0-only

package sdk

import (
	"testing"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		version string
		want    [3]int
		wantErr bool
	}{
		{"2.0.0", [3]int{2, 0, 0}, false},
		{"2.1.5", [3]int{2, 1, 5}, false},
		{"1.0.0", [3]int{1, 0, 0}, false},
		{"invalid", [3]int{}, true},
		{"2.0", [3]int{}, true},
		{"", [3]int{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			got, err := parseVersion(tt.version)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"2.0.0", "2.0.0", 0},
		{"2.1.0", "2.0.0", 1},
		{"2.0.0", "2.1.0", -1},
		{"3.0.0", "2.9.9", 1},
		{"1.9.9", "2.0.0", -1},
	}

	for _, tt := range tests {
		t.Run(tt.a+"_vs_"+tt.b, func(t *testing.T) {
			got := compareVersions(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestCheckMinimumVersion(t *testing.T) {
	tests := []struct {
		version string
		wantErr bool
	}{
		{"2.0.0", false},
		{"2.1.0", false},
		{"3.0.0", false},
		{"1.9.9", true},
		{"1.0.0", true},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			err := checkMinimumVersion(tt.version)
			if tt.wantErr && err == nil {
				t.Error("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
