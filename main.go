package main

import (
	"os"

	"github.com/benjih/kpass/internal/args"
	"github.com/benjih/kpass/internal/kde"
	"github.com/benjih/kpass/internal/runtime"
	qt "github.com/mappu/miqt/qt6"
)

func init() {
	// Qt reads QT_PLUGIN_PATH and XDG_DATA_DIRS before QApplication starts.
	runtime.SetupRuntimeEnvironment()
}

func main() {
	qt.NewQApplication(os.Args)
	kde.SetupKDEAppearance()
	filePath, fillHost := args.ParseInitialArg(os.Args[1:])
	kde.SetupUI(filePath, fillHost)
	qt.QApplication_Exec()
}
