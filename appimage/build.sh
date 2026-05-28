#!/bin/bash
# Builds AppDir/ from a compiled kpass binary, then packages it as an AppImage.
# Run from the repository root after: go build -tags appimage -o kpass .
#
# Required env:
#   QMAKE   path to qmake6 (used by linuxdeploy-plugin-qt to find Qt)
#
# Optional:
#   ARCH            target architecture string (default: x86_64)
#   OUTPUT          output filename (default: KPass-$ARCH.AppImage)
#   APPDIR_PATH     where to stage the AppDir (default: AppDir)
#
# Tools resolved from PATH, or as AppImages placed in appimage/:
#   linuxdeploy, linuxdeploy-plugin-qt, appimagetool

set -euo pipefail

ARCH="${ARCH:-x86_64}"
APPDIR="${APPDIR_PATH:-AppDir}"
OUTPUT="${OUTPUT:-KPass-${ARCH}.AppImage}"
REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"

# ---------------------------------------------------------------------------
# Resolve tools — prefer AppImages dropped in appimage/, else PATH
# ---------------------------------------------------------------------------
resolve_tool() {
    local name="$1"
    local appimage="${REPO_ROOT}/appimage/${name}-${ARCH}.AppImage"
    if [ -x "$appimage" ]; then
        echo "$appimage"
    else
        command -v "$name" 2>/dev/null || { echo "ERROR: $name not found in PATH or appimage/" >&2; exit 1; }
    fi
}

LINUXDEPLOY="$(resolve_tool linuxdeploy)"
LINUXDEPLOY_PLUGIN_QT="$(resolve_tool linuxdeploy-plugin-qt)"
APPIMAGETOOL="$(resolve_tool appimagetool)"

# ---------------------------------------------------------------------------
# Validate inputs
# ---------------------------------------------------------------------------
[ -x "$REPO_ROOT/kpass" ] || { echo "ERROR: $REPO_ROOT/kpass not found — run: go build -tags appimage -o kpass ." >&2; exit 1; }
[ -n "${QMAKE:-}" ] || { echo "ERROR: set QMAKE to the path of qmake6" >&2; exit 1; }

QML_BASE="$("$QMAKE" -query QT_INSTALL_QML)"

# ---------------------------------------------------------------------------
# Phase 1: Scaffold AppDir
# ---------------------------------------------------------------------------
rm -rf "$APPDIR"
mkdir -p \
    "$APPDIR/usr/bin" \
    "$APPDIR/usr/share/kpass/assets" \
    "$APPDIR/usr/share/applications" \
    "$APPDIR/usr/share/icons/hicolor/512x512/apps" \
    "$APPDIR/usr/share/metainfo"

cp "$REPO_ROOT/kpass"                                    "$APPDIR/usr/bin/kpass"
cp -a "$REPO_ROOT/qml"                                   "$APPDIR/usr/share/kpass/qml"
cp "$REPO_ROOT/assets/KPass.png"                         "$APPDIR/usr/share/kpass/assets/KPass.png"
cp "$REPO_ROOT/assets/KPass.png"                         "$APPDIR/usr/share/icons/hicolor/512x512/apps/com.benjih.KPass.png"
cp "$REPO_ROOT/assets/KPass.png"                         "$APPDIR/com.benjih.KPass.png"
cp "$REPO_ROOT/flatpak/com.benjih.KPass.desktop"         "$APPDIR/usr/share/applications/com.benjih.KPass.desktop"
cp "$REPO_ROOT/flatpak/com.benjih.KPass.desktop"         "$APPDIR/com.benjih.KPass.desktop"
cp "$REPO_ROOT/flatpak/com.benjih.KPass.metainfo.xml"    "$APPDIR/usr/share/metainfo/com.benjih.KPass.metainfo.xml"
cp "$REPO_ROOT/appimage/AppRun"                          "$APPDIR/AppRun"
chmod +x "$APPDIR/AppRun"

# ---------------------------------------------------------------------------
# Phase 2: Bundle Qt libraries and plugins via linuxdeploy + plugin-qt
# linuxdeploy-plugin-qt scans qml/ for Qt QML imports via qmlimportscanner.
# We do NOT pass --output here — linuxdeploy only deploys into AppDir.
# ---------------------------------------------------------------------------
export QMAKE
export QMLDIR="$REPO_ROOT/qml"

LINUXDEPLOY_PLUGIN_QT="$LINUXDEPLOY_PLUGIN_QT" \
"$LINUXDEPLOY" \
    --appdir "$APPDIR" \
    --executable "$APPDIR/usr/bin/kpass" \
    --desktop-file "$APPDIR/com.benjih.KPass.desktop" \
    --icon-file "$APPDIR/com.benjih.KPass.png" \
    --plugin qt

# ---------------------------------------------------------------------------
# Phase 3: Bundle KDE QML modules not caught by linuxdeploy-plugin-qt.
# linuxdeploy-plugin-qt only deploys Qt's own QML sources; KDE Frameworks
# QML plugins (Kirigami, QQC2 desktop style, etc.) must be copied manually.
# ---------------------------------------------------------------------------
KDE_QML_MODULES=(
    "org/kde/kirigami"
    "org/kde/qqc2desktopstyle"
    "org/kde/kquickcontrolsaddons"
    "org/kde/kitemmodels"
)

APPDIR_QML="$APPDIR/usr/qml"
mkdir -p "$APPDIR_QML"

for mod in "${KDE_QML_MODULES[@]}"; do
    src="$QML_BASE/$mod"
    if [ -d "$src" ]; then
        mkdir -p "$APPDIR_QML/$(dirname "$mod")"
        cp -a "$src" "$APPDIR_QML/$(dirname "$mod")/"
        echo "Bundled KDE QML module: $mod"
    else
        echo "WARNING: KDE QML module not found at $src" >&2
    fi
done

# ---------------------------------------------------------------------------
# Phase 4: Package with appimagetool
# ---------------------------------------------------------------------------
ARCH="$ARCH" "$APPIMAGETOOL" "$APPDIR" "$OUTPUT"

echo ""
echo "AppImage created: $OUTPUT"
