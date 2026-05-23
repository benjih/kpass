package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/benjih/kpass/bridge"
	qt "github.com/mappu/miqt/qt6"
	"github.com/mappu/miqt/qt6/qml"
)

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
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}
