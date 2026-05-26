//go:build flatpak

package main

import "os"

func setupRuntimeEnvironment() {
	if os.Getenv("QT_QUICK_CONTROLS_STYLE") == "" {
		os.Setenv("QT_QUICK_CONTROLS_STYLE", "org.kde.desktop")
	}
	if os.Getenv("XDG_ICON_THEME") == "" {
		os.Setenv("XDG_ICON_THEME", "breeze")
	}
	if os.Getenv("QT_QPA_PLATFORMTHEME") == "" {
		os.Setenv("QT_QPA_PLATFORMTHEME", "kde")
	}
}
