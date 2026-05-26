package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/benjih/kpass/internal"
)

var (
	reQtBaseStore = regexp.MustCompile(`-qtbase-6\.(\d+)\.(\d+)$`)
	reQtDeclStore = regexp.MustCompile(`-qtdeclarative-6\.(\d+)\.(\d+)$`)
	reLibGLStore  = regexp.MustCompile(`-libglvnd-(\d+)\.(\d+)\.(\d+)$`)
)

// setupRuntimeEnvironment ensures devbox Qt/KDE paths work when the binary is
// launched without sourcing scripts/kde-env.sh (e.g. ./kpass after make build).
func setupRuntimeEnvironment() {
	if os.Getenv("FLATPAK_ID") != "" {
		if os.Getenv("QT_QUICK_CONTROLS_STYLE") == "" {
			os.Setenv("QT_QUICK_CONTROLS_STYLE", "org.kde.desktop")
		}
		if os.Getenv("XDG_ICON_THEME") == "" {
			os.Setenv("XDG_ICON_THEME", "breeze")
		}
		if os.Getenv("QT_QPA_PLATFORMTHEME") == "" {
			os.Setenv("QT_QPA_PLATFORMTHEME", "kde")
		}
		return
	}

	root := internal.ProjectRoot()
	profile := filepath.Join(root, ".devbox/nix/profile/default")
	closure := internal.NixClosurePaths(profile)

	pluginPaths := kdePluginPaths()
	internal.AugmentEnvPath("XDG_DATA_DIRS", iconSharePaths()...)
	internal.AugmentEnvPath("QT_PLUGIN_PATH", pluginPaths...)
	internal.AugmentEnvPathFiltered("LD_LIBRARY_PATH", []string{"qtbase-", "qtdeclarative-"}, qtLibraryPaths(closure)...)
	internal.AugmentEnvPath("QML2_IMPORT_PATH", qmlImportPaths(root, closure)...)

	if os.Getenv("QT_QPA_PLATFORMTHEME") == "" && len(pluginPaths) > 0 {
		os.Setenv("QT_QPA_PLATFORMTHEME", "kde")
	}
	if os.Getenv("XDG_ICON_THEME") == "" {
		os.Setenv("XDG_ICON_THEME", "breeze")
	}
	if os.Getenv("QT_QUICK_CONTROLS_STYLE") == "" {
		os.Setenv("QT_QUICK_CONTROLS_STYLE", "org.kde.desktop")
	}
	// Nix devbox closures often lack a working GLX stack for Qt Quick's default RHI path.
	if os.Getenv("QT_QUICK_BACKEND") == "" {
		os.Setenv("QT_QUICK_BACKEND", "software")
	}
}

func qtLibraryPaths(closure []string) []string {
	var paths []string
	if base := newestStoreMatch(closure, reQtBaseStore); base != "" {
		paths = append(paths, filepath.Join(base, "lib"))
	}
	if decl := newestStoreMatch(closure, reQtDeclStore); decl != "" {
		paths = append(paths, filepath.Join(decl, "lib"))
	}
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
