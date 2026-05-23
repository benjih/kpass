.PHONY: run build generate moc

run: generate
	. scripts/kde-env.sh && QT_QUICK_BACKEND=software go run .

build: generate
	. scripts/kde-env.sh && go build -o kpass .

generate: bridge/moc_databasemanager.cpp

# Qt 6.7 moc from KDE closure (see scripts/kde-env.sh).
MOC ?= $(shell . scripts/kde-env.sh 2>/dev/null; \
	command -v moc 2>/dev/null || \
	ls $${QTBASE:-/nonexistent}/libexec/moc 2>/dev/null || \
	ls .devbox/nix/profile/default/libexec/moc 2>/dev/null)

bridge/moc_databasemanager.cpp: bridge/databasemanager.h
	@test -n "$(MOC)" && test -x "$(MOC)" || (echo "Run 'devbox install' inside devbox shell, then make generate" && exit 1)
	$(MOC) bridge/databasemanager.h -o bridge/moc_databasemanager.cpp
