package version

import (
	"runtime/debug"
	"testing"
)

func TestFromSettings(t *testing.T) {
	tests := []struct {
		name     string
		settings []debug.BuildSetting
		want     string
	}{
		{
			name:     "no vcs info",
			settings: nil,
			want:     Dev,
		},
		{
			name: "clean build",
			settings: []debug.BuildSetting{
				{Key: "vcs", Value: "git"},
				{Key: "vcs.revision", Value: "5140483922f222e607cbc4d475afc635e8eb8b60"},
				{Key: "vcs.modified", Value: "false"},
			},
			want: "5140483",
		},
		{
			name: "dirty build",
			settings: []debug.BuildSetting{
				{Key: "vcs.revision", Value: "5140483922f222e607cbc4d475afc635e8eb8b60"},
				{Key: "vcs.modified", Value: "true"},
			},
			want: "5140483-dirty",
		},
		{
			name: "short revision kept as-is",
			settings: []debug.BuildSetting{
				{Key: "vcs.revision", Value: "abc"},
			},
			want: "abc",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fromSettings(tt.settings); got != tt.want {
				t.Errorf("fromSettings() = %q, want %q", got, tt.want)
			}
		})
	}
}
