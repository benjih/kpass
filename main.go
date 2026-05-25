package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/benjih/kpass/bridge"
	qt "github.com/mappu/miqt/qt6"
	"github.com/mappu/miqt/qt6/qml"
)

func init() {
	// Qt reads QT_PLUGIN_PATH and XDG_DATA_DIRS before QApplication starts.
	setupRuntimeEnvironment()
}

func main() {
	qt.NewQApplication(os.Args)
	setupKDEAppearance()

	engine := qml.NewQQmlApplicationEngine()
	engine.RootContext().SetContextProperty("databaseManager", bridge.NewDatabaseManager())

	root := projectRoot()
	engine.RootContext().SetContextProperty2(
		"appIconUrl",
		qt.NewQVariant11(qt.QUrl_FromLocalFile(filepath.Join(root, "assets", "KPass.png")).ToString()),
	)

	qmlDir := filepath.Join(root, "qml")
	mainQML := filepath.Join(qmlDir, "Main.qml")
	engine.AddImportPath(qmlDir)
	for _, importPath := range strings.Split(os.Getenv("QML2_IMPORT_PATH"), ":") {
		if importPath != "" {
			engine.AddImportPath(importPath)
		}
	}
	engine.Load(qt.QUrl_FromLocalFile(mainQML))

	if len(engine.RootObjects()) == 0 {
		os.Exit(1)
	}

	qt.QApplication_Exec()
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

// setupKDEAppearance mirrors kpass-c/src/main.cpp so QSettings, icons, and QQC2 match KDE apps.
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

func projectRoot() string {
	if exe, err := os.Executable(); err == nil {
		root := filepath.Dir(exe)
		if !strings.Contains(root, "go-build") {
			if _, err := os.Stat(filepath.Join(root, "qml", "Main.qml")); err == nil {
				return root
			}
		}
	}
	if wd, err := os.Getwd(); err == nil {
		return wd
	}
	return "."
}
