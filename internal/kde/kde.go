package kde

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/benjih/kpass/internal"
	"github.com/benjih/kpass/internal/bridge"
	qt "github.com/mappu/miqt/qt6"
	"github.com/mappu/miqt/qt6/qml"
)

func SetupKDEAppearance() {
	qt.QCoreApplication_SetApplicationName("KPass")
	qt.QCoreApplication_SetOrganizationName("KDE")
	qt.QCoreApplication_SetOrganizationDomain("kde.org")

	qt.QIcon_SetThemeName("breeze")
	qt.QIcon_SetFallbackThemeName("breeze")
	if extra := iconThemeSearchPaths(); len(extra) > 0 {
		paths := append(extra, qt.QIcon_ThemeSearchPaths()...)
		qt.QIcon_SetThemeSearchPaths(paths)
	}

	iconPath := filepath.Join(internal.ProjectRoot(), "assets", "KPass.png")
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

func KdePluginPaths() []string {
	root := internal.ProjectRoot()
	profile := filepath.Join(root, ".devbox/nix/profile/default")

	var paths []string
	addPlugins := func(base string) {
		p := filepath.Join(base, "lib", "qt-6", "plugins")
		if info, err := os.Stat(p); err == nil && info.IsDir() {
			paths = append(paths, p)
		}
	}

	addPlugins(profile)
	for _, store := range internal.NixClosurePaths(profile) {
		addPlugins(store)
	}
	return paths
}

func IconSharePaths() []string {
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
	add(filepath.Join(internal.ProjectRoot(), ".devbox/nix/profile/default/share"))

	for _, dir := range strings.Split(os.Getenv("XDG_DATA_DIRS"), ":") {
		add(dir)
	}
	return paths
}

func SetupUI(initialFilePath, initialFillHost string) {
	engine := qml.NewQQmlApplicationEngine()
	engine.RootContext().SetContextProperty("databaseManager", bridge.NewDatabaseManager())

	root := internal.ProjectRoot()
	engine.RootContext().SetContextProperty2(
		"appIconUrl",
		qt.NewQVariant11(qt.QUrl_FromLocalFile(filepath.Join(root, "assets", "KPass.png")).ToString()),
	)
	engine.RootContext().SetContextProperty2(
		"initialFilePath",
		qt.NewQVariant11(initialFilePath),
	)
	engine.RootContext().SetContextProperty2(
		"initialFillHost",
		qt.NewQVariant11(initialFillHost),
	)

	qmlDir := filepath.Join(root, "qml")
	engine.AddImportPath(qmlDir)
	for _, importPath := range strings.Split(os.Getenv("QML2_IMPORT_PATH"), ":") {
		if importPath != "" {
			engine.AddImportPath(importPath)
		}
	}
	engine.Load(qt.QUrl_FromLocalFile(filepath.Join(qmlDir, "Main.qml")))

	if len(engine.RootObjects()) == 0 {
		os.Exit(1)
	}
}
