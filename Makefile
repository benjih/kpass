.PHONY: run build generate moc

run: generate
	QT_QUICK_BACKEND=software go run .

build: generate
	go build -o kpass .

generate: bridge/moc_databasemanager.cpp

MOC ?= .devbox/nix/profile/default/libexec/moc

bridge/moc_databasemanager.cpp: bridge/databasemanager.h
	@test -x "$(MOC)" || (echo "Run 'devbox install' and use 'devbox run make generate'" && exit 1)
	$(MOC) bridge/databasemanager.h -o bridge/moc_databasemanager.cpp
