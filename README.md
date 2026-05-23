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

Use `QT_QUICK_BACKEND=software` if the GPU backend fails (already set in the Makefile `run` target).

## Notes

- QML uses **Qt Quick Controls**, not Kirigami (Kirigami is not in the devbox Qt package). Behavior matches `kpass-c` where possible.
- Run from the project directory so `qml/Main.qml` resolves.
- `bridge/moc_databasemanager.cpp` is generated locally and gitignored.
