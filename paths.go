package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// projectRoot resolves the directory containing QML assets. It probes the
// executable's location first (skipping go-build temp dirs from `go run`),
// then checks the FHS-style share/kpass sibling for installed/Flatpak layouts,
// and falls back to the working directory for development.
func projectRoot() string {
	if exe, err := os.Executable(); err == nil {
		binDir := filepath.Dir(exe)
		if !strings.Contains(binDir, "go-build") {
			for _, root := range []string{
				binDir,
				filepath.Join(binDir, "..", "share", "kpass"),
			} {
				if _, err := os.Stat(filepath.Join(root, "qml", "Main.qml")); err == nil {
					return root
				}
			}
		}
	}
	if wd, err := os.Getwd(); err == nil {
		return wd
	}
	return "."
}

func nixClosurePaths(profile string) []string {
	if _, err := os.Stat(profile); err != nil {
		return nil
	}
	out, err := exec.Command("nix-store", "-qR", profile).Output()
	if err != nil {
		return nil
	}

	var paths []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			paths = append(paths, line)
		}
	}
	return paths
}

func augmentEnvPath(key string, paths ...string) {
	augmentEnvPathFiltered(key, nil, paths...)
}

func augmentEnvPathFiltered(key string, skipSubstrings []string, paths ...string) {
	if len(paths) == 0 && len(skipSubstrings) == 0 {
		return
	}

	shouldSkip := func(p string) bool {
		for _, sub := range skipSubstrings {
			if sub != "" && strings.Contains(p, sub) {
				return true
			}
		}
		return false
	}

	seen := map[string]bool{}
	var merged []string
	add := func(p string) {
		if p == "" || seen[p] || shouldSkip(p) {
			return
		}
		info, err := os.Stat(p)
		if err != nil || !info.IsDir() {
			return
		}
		seen[p] = true
		merged = append(merged, p)
	}

	for _, p := range paths {
		add(p)
	}
	for _, p := range strings.Split(os.Getenv(key), ":") {
		add(p)
	}
	if len(merged) > 0 {
		os.Setenv(key, strings.Join(merged, ":"))
	}
}
