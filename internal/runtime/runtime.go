//go:build !flatpak && !appimage

package runtime

import (
	"os"
	"path/filepath"

	"github.com/benjih/kpass/internal"
	"github.com/benjih/kpass/internal/kde"
)

// SetupRuntimeEnvironment ensures devbox Qt/KDE paths work when the binary is
// launched without sourcing scripts/kde-env.sh (e.g. ./kpass after make build).
func SetupRuntimeEnvironment() {
	root := internal.ProjectRoot()
	profile := filepath.Join(root, ".devbox/nix/profile/default")
	closure := internal.NixClosurePaths(profile)

	pluginPaths := kde.KdePluginPaths()
	internal.AugmentEnvPath("XDG_DATA_DIRS", kde.IconSharePaths()...)
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
