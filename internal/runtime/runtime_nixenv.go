//go:build !flatpak && !appimage

package runtime

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func qtLibraryPaths(closure []string) []string {
	var paths []string
	reQtBaseStore := regexp.MustCompile(`-qtbase-6\.(\d+)\.(\d+)$`)
	if base := newestStoreMatch(closure, reQtBaseStore); base != "" {
		paths = append(paths, filepath.Join(base, "lib"))
	}

	reQtDeclStore := regexp.MustCompile(`-qtdeclarative-6\.(\d+)\.(\d+)$`)
	if decl := newestStoreMatch(closure, reQtDeclStore); decl != "" {
		paths = append(paths, filepath.Join(decl, "lib"))
	}

	reLibGLStore := regexp.MustCompile(`-libglvnd-(\d+)\.(\d+)\.(\d+)$`)
	if gl := newestStoreMatch(closure, reLibGLStore); gl != "" {
		paths = append(paths, filepath.Join(gl, "lib"))
	}
	if info, err := os.Stat("/run/opengl-driver/lib"); err == nil && info.IsDir() {
		paths = append(paths, "/run/opengl-driver/lib")
	}
	return paths
}

func qmlImportPaths(root string, closure []string) []string {
	var paths []string
	seen := map[string]bool{}
	add := func(p string) {
		if p == "" || seen[p] {
			return
		}
		if info, err := os.Stat(p); err == nil && info.IsDir() {
			seen[p] = true
			paths = append(paths, p)
		}
	}

	add(filepath.Join(root, ".devbox/nix/profile/default/lib/qt-6/qml"))
	for _, store := range closure {
		add(filepath.Join(store, "lib/qt-6/qml"))
	}
	return paths
}

func newestStoreMatch(paths []string, re *regexp.Regexp) string {
	var best string
	var bestVer [3]int
	for _, p := range paths {
		base := filepath.Base(p)
		if strings.Contains(base, "-dev") || strings.Contains(base, "only-plugins") {
			continue
		}
		m := re.FindStringSubmatch(base)
		if m == nil {
			continue
		}
		ver := parseVersionInts(m[1:])
		if versionGreater(ver, bestVer) {
			best = p
			bestVer = ver
		}
	}
	return best
}

func parseVersionInts(parts []string) [3]int {
	var ver [3]int
	for i := 0; i < len(parts) && i < 3; i++ {
		ver[i], _ = strconv.Atoi(parts[i])
	}
	return ver
}

func versionGreater(a, b [3]int) bool {
	for i := 0; i < 3; i++ {
		if a[i] != b[i] {
			return a[i] > b[i]
		}
	}
	return false
}
