package main

import (
	"os"
	"strings"

	qt "github.com/mappu/miqt/qt6"
)

func init() {
	// Qt reads QT_PLUGIN_PATH and XDG_DATA_DIRS before QApplication starts.
	setupRuntimeEnvironment()
}

func main() {
	qt.NewQApplication(os.Args)
	setupKDEAppearance()
	setupUI(initialFileArg())
	qt.QApplication_Exec()
}

func initialFileArg() string {
	for _, arg := range os.Args[1:] {
		if !strings.HasPrefix(arg, "-") {
			return arg
		}
	}
	return ""
}
