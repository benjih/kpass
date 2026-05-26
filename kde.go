package main

import (
	"os"
	"path/filepath"
	"strings"

	qt "github.com/mappu/miqt/qt6"
)

func setupKDEAppearance() {
	qt.QCoreApplication_SetApplicationName("KPass")
	qt.QCoreApplication_SetOrganizationName("KDE")
	qt.QCoreApplication_SetOrganizationDomain("kde.org")

	qt.QIcon_SetThemeName("breeze")
	qt.QIcon_SetFallbackThemeName("breeze")
	if extra := iconThemeSearchPaths(); len(extra) > 0 {
		paths := append(extra, qt.QIcon_ThemeSearchPaths()...)
		qt.QIcon_SetThemeSearchPaths(paths)
	}

	iconPath := filepath.Join(projectRoot(), "assets", "KPass.png")
	if _, err := os.Stat(iconPath); err == nil {
		qt.QGuiApplication_SetWindowIcon(qt.NewQIcon4(iconPath))
	}
}

// iconThemeSearchPaths returns deduplicated icon theme directories. In dev it
// falls back to the devbox nix profile share directory; KPASS_ICON_PROFILE
// overrides that (e.g. for CI). XDG_DATA_DIRS covers installed system layouts.
// Only directories with an `icons` subdirectory are included.
func iconThemeSearchPaths() []string {
	var paths []string
	seen := map[string]bool{}
	add := func(p string) {
		if p == "" || seen[p] {
			return
		}
		icons := filepath.Join(p, "icons")
		if info, err := os.Stat(icons); err == nil && info.IsDir() {
			seen[icons] = true
			paths = append(paths, icons)
		}
	}

	if profile := os.Getenv("KPASS_ICON_PROFILE"); profile != "" {
		add(profile)
	} else if wd, err := os.Getwd(); err == nil {
		add(filepath.Join(wd, ".devbox/nix/profile/default/share"))
	}

	for _, dir := range strings.Split(os.Getenv("XDG_DATA_DIRS"), ":") {
		add(dir)
	}
	return paths
}

func kdePluginPaths() []string {
	root := projectRoot()
	profile := filepath.Join(root, ".devbox/nix/profile/default")

	var paths []string
	addPlugins := func(base string) {
		p := filepath.Join(base, "lib", "qt-6", "plugins")
		if info, err := os.Stat(p); err == nil && info.IsDir() {
			paths = append(paths, p)
		}
	}

	addPlugins(profile)
	for _, store := range nixClosurePaths(profile) {
		addPlugins(store)
	}
	return paths
}

func iconSharePaths() []string {
	var paths []string
	seen := map[string]bool{}
	add := func(share string) {
		if share == "" || seen[share] {
			return
		}
		if info, err := os.Stat(share); err == nil && info.IsDir() {
			seen[share] = true
			paths = append(paths, share)
		}
	}

	if profile := os.Getenv("KPASS_ICON_PROFILE"); profile != "" {
		add(profile)
	}
	add(filepath.Join(projectRoot(), ".devbox/nix/profile/default/share"))

	for _, dir := range strings.Split(os.Getenv("XDG_DATA_DIRS"), ":") {
		add(dir)
	}
	return paths
}
