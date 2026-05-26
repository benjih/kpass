package main

import (
	"os"

	qt "github.com/mappu/miqt/qt6"
)

func init() {
	// Qt reads QT_PLUGIN_PATH and XDG_DATA_DIRS before QApplication starts.
	setupRuntimeEnvironment()
}

func main() {
	qt.NewQApplication(os.Args)
	setupKDEAppearance()
	setupUI()
	qt.QApplication_Exec()
}
