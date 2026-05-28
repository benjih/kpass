.PHONY: run build generate moc \
	flatpak flatpak-deps flatpak-vendor flatpak-build flatpak-install flatpak-run flatpak-rebuild flatpak-clean \
	appimage appimage-deps appimage-build appimage-appdir appimage-clean

FLATPAK_MANIFEST := flatpak/com.benjih.KPass.yaml
FLATPAK_ID := com.benjih.KPass
FLATPAK_REPO := repo
FLATPAK_REPO_ABS := $(abspath $(FLATPAK_REPO))
FLATPAK_BUILD_DIR := flatpak-build
FLATPAK_REMOTE := flathub
FLATPAK_RUNTIME_VERSION := 6.10
FLATPAK_GOLANG_EXT_VERSION := 25.08

run: generate
	. scripts/kde-env.sh && go run .

build: generate
	. scripts/kde-env.sh && go build -o kpass .

generate: internal/bridge/moc_databasemanager.cpp

# Qt moc from devbox closure (see scripts/kde-env.sh). Prefer nix Qt over /usr/bin/moc.
MOC ?= $(shell . scripts/kde-env.sh 2>/dev/null; \
	if [ -n "$$QTBASE" ] && [ -x "$$QTBASE/libexec/moc" ]; then echo "$$QTBASE/libexec/moc"; \
	elif [ -n "$$QTBASE" ] && [ -x "$$QTBASE/bin/moc" ]; then echo "$$QTBASE/bin/moc"; \
	else command -v moc 2>/dev/null; fi)

internal/bridge/moc_databasemanager.cpp: internal/bridge/databasemanager.h
	@test -n "$(MOC)" && test -x "$(MOC)" || (echo "Run 'devbox install' inside devbox shell, then make generate" && exit 1)
	$(MOC) internal/bridge/databasemanager.h -o internal/bridge/moc_databasemanager.cpp

# Flatpak: vendor deps, install SDK/runtime, build, and install the app.
flatpak: flatpak-deps flatpak-vendor flatpak-build flatpak-install
	@echo "Installed $(FLATPAK_ID) — run: make flatpak-run"

flatpak-deps:
	@command -v flatpak >/dev/null || (echo "Install flatpak" && exit 1)
	@command -v flatpak-builder >/dev/null || (echo "Install flatpak-builder" && exit 1)
	flatpak install -y $(FLATPAK_REMOTE) \
		org.kde.Platform//$(FLATPAK_RUNTIME_VERSION) \
		org.kde.Sdk//$(FLATPAK_RUNTIME_VERSION) \
		org.freedesktop.Sdk.Extension.golang//$(FLATPAK_GOLANG_EXT_VERSION)

flatpak-vendor: vendor/modules.txt

vendor/modules.txt: go.mod go.sum
	rm -rf vendor
	go mod vendor

flatpak-build: flatpak-vendor $(FLATPAK_MANIFEST)
	@if [ -d $(FLATPAK_BUILD_DIR) ] && [ ! -f $(FLATPAK_BUILD_DIR)/metadata ]; then \
		echo "Removing incomplete $(FLATPAK_BUILD_DIR) from a previous failed build"; \
		rm -rf $(FLATPAK_BUILD_DIR); \
	fi
	flatpak-builder --install-deps-from=$(FLATPAK_REMOTE) \
		--repo=$(FLATPAK_REPO) \
		$(FLATPAK_BUILD_DIR) \
		$(FLATPAK_MANIFEST) \
		$(FLATPAK_BUILD_FLAGS) \
		--force-clean

flatpak-install:
	@test -d "$(FLATPAK_REPO_ABS)" || (echo "Run make flatpak-build first (missing $(FLATPAK_REPO_ABS))" && exit 1)
	flatpak --user install -y --reinstall "$(FLATPAK_REPO_ABS)" $(FLATPAK_ID) 2>/dev/null \
		|| flatpak install -y --reinstall "$(FLATPAK_REPO_ABS)" $(FLATPAK_ID)

flatpak-run:
	flatpak run $(FLATPAK_ID)

flatpak-clean:
	rm -rf $(FLATPAK_BUILD_DIR) $(FLATPAK_REPO)

flatpak-rebuild: flatpak-clean
	$(MAKE) flatpak

# AppImage: build binary, assemble AppDir, and package with appimagetool.
# Requires: QMAKE env var pointing to qmake6, and linuxdeploy / linuxdeploy-plugin-qt /
# appimagetool either in PATH or as AppImages placed in appimage/.
ARCH ?= x86_64
APPIMAGE_OUTPUT ?= KPass-$(ARCH).AppImage

APPIMAGE_TOOLS_DIR := appimage
LINUXDEPLOY_BASE_URL := https://github.com/linuxdeploy/linuxdeploy/releases/download/continuous
LINUXDEPLOY_QT_BASE_URL := https://github.com/linuxdeploy/linuxdeploy-plugin-qt/releases/download/continuous
APPIMAGETOOL_BASE_URL := https://github.com/AppImage/appimagetool/releases/download/continuous

appimage: appimage-deps appimage-build appimage-appdir

appimage-deps:
	@command -v wget >/dev/null || (echo "Install wget" && exit 1)
	wget -qNP $(APPIMAGE_TOOLS_DIR) $(LINUXDEPLOY_BASE_URL)/linuxdeploy-$(ARCH).AppImage
	wget -qNP $(APPIMAGE_TOOLS_DIR) $(LINUXDEPLOY_QT_BASE_URL)/linuxdeploy-plugin-qt-$(ARCH).AppImage
	wget -qNP $(APPIMAGE_TOOLS_DIR) $(APPIMAGETOOL_BASE_URL)/appimagetool-$(ARCH).AppImage
	chmod +x $(APPIMAGE_TOOLS_DIR)/linuxdeploy-$(ARCH).AppImage \
		$(APPIMAGE_TOOLS_DIR)/linuxdeploy-plugin-qt-$(ARCH).AppImage \
		$(APPIMAGE_TOOLS_DIR)/appimagetool-$(ARCH).AppImage

appimage-build: generate
	go build -tags appimage -trimpath -ldflags="-s -w" -o kpass .

appimage-appdir: appimage-build
	bash appimage/build.sh

appimage-clean:
	rm -rf AppDir $(APPIMAGE_OUTPUT)
