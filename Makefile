.PHONY: run build generate moc

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
