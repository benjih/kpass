//go:build appimage

package runtime

import (
	"os"
	"path/filepath"

	"github.com/benjih/kpass/internal"
)

func SetupRuntimeEnvironment() {
	appdir := os.Getenv("APPDIR")
	if appdir == "" {
		if exe, err := os.Executable(); err == nil {
			// binary lives at AppDir/usr/bin/kpass → go up three levels
			appdir = filepath.Join(filepath.Dir(exe), "..", "..", "..")
		}
	}

	internal.AugmentEnvPath("LD_LIBRARY_PATH", filepath.Join(appdir, "usr/lib"))
	internal.AugmentEnvPath("QT_PLUGIN_PATH", filepath.Join(appdir, "usr/plugins"))
	internal.AugmentEnvPath("QML2_IMPORT_PATH", filepath.Join(appdir, "usr/qml"))
	internal.AugmentEnvPath("XDG_DATA_DIRS", filepath.Join(appdir, "usr/share"))

	if os.Getenv("QT_QPA_PLATFORMTHEME") == "" {
		os.Setenv("QT_QPA_PLATFORMTHEME", "kde")
	}
	if os.Getenv("QT_QUICK_CONTROLS_STYLE") == "" {
		os.Setenv("QT_QUICK_CONTROLS_STYLE", "org.kde.desktop")
	}
	if os.Getenv("XDG_ICON_THEME") == "" {
		os.Setenv("XDG_ICON_THEME", "breeze")
	}
}
