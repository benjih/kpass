.PHONY: run build generate moc \
	flatpak flatpak-deps flatpak-vendor flatpak-build flatpak-install flatpak-run flatpak-rebuild flatpak-clean

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

generate: bridge/moc_databasemanager.cpp

# Qt moc from devbox closure (see scripts/kde-env.sh). Prefer nix Qt over /usr/bin/moc.
MOC ?= $(shell . scripts/kde-env.sh 2>/dev/null; \
	if [ -n "$$QTBASE" ] && [ -x "$$QTBASE/libexec/moc" ]; then echo "$$QTBASE/libexec/moc"; \
	elif [ -n "$$QTBASE" ] && [ -x "$$QTBASE/bin/moc" ]; then echo "$$QTBASE/bin/moc"; \
	else command -v moc 2>/dev/null; fi)

bridge/moc_databasemanager.cpp: bridge/databasemanager.h
	@test -n "$(MOC)" && test -x "$(MOC)" || (echo "Run 'devbox install' inside devbox shell, then make generate" && exit 1)
	$(MOC) bridge/databasemanager.h -o bridge/moc_databasemanager.cpp

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
	$(MAKE) flatpak-deps flatpak-vendor flatpak-build flatpak-install
	@echo "Installed $(FLATPAK_ID) — run: make flatpak-run"
