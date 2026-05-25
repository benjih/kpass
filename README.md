# KPass (Go + Qt QML)

Desktop KeePass client combining:

- **Go + [miqt](https://github.com/mappu/miqt)** for Qt 6 / QML (from the hello-world prototype)
- **[gokeepasslib](https://github.com/tobischo/gokeepasslib)** for `.kdbx` I/O (same library as `password/`)
- **QML UI** ported from `kpass-c/` (splash, entry list, details, group filter)
- **Thin C++ `DatabaseManager` bridge** so QML can call `openDatabase`, `saveDatabase`, etc. (miqt cannot expose new `Q_INVOKABLE` methods from Go alone)

## Layout

```
internal/keepass/   KeePass open/save/traverse logic (Go)
bridge/             C++ QObject + cgo exports for QML
qml/                Qt Quick Controls UI (ported from Kirigami version)
main.go             Application entry
```

## Run

```bash
devbox shell
make generate   # once: runs Qt 6 moc on bridge/databasemanager.h
make run        # or: go run .
```

`QT_QUICK_BACKEND=software` is the default (in `scripts/kde-env.sh` and at app startup) so `./kpass` works on Nix without a full GLX stack. Set `QT_QUICK_BACKEND=basic` to try GPU rendering.

## KDE / Kirigami

Devbox includes Kirigami, Breeze icons, and Breeze desktop QQC2 style (pinned to **6.18.0**). Those packages pull in a matching Qt (currently **6.9**); do not add `qt6.full` or another Qt pin, or you will get “incompatible Qt library” errors from KDE QML plugins.

`scripts/kde-env.sh` discovers Qt from the devbox closure and sets `PKG_CONFIG_PATH`, `LD_LIBRARY_PATH`, and `QML2_IMPORT_PATH` for `go build` / `go run`, plus `QT_QUICK_CONTROLS_STYLE=org.kde.desktop`.

After changing Qt packages, run: `go clean -cache && make build`.

User-visible strings use the `Tr` QML singleton (`Tr.i18n`) instead of KI18n.

QML is the original `kpass-c` Kirigami UI with Qt 6 import syntax and a Qt 6 `FileDialog` on the splash screen.

## Notes

- Run from the project directory so `qml/Main.qml` resolves.
- `bridge/moc_databasemanager.cpp` is generated locally and gitignored.
- Use `devbox run make build` so KDE paths and CGO flags are set.
