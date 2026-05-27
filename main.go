package main

import (
	"net/url"
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
	filePath, fillHost := parseInitialArg()
	setupUI(filePath, fillHost)
	qt.QApplication_Exec()
}

// parseInitialArg inspects the first non-flag argument and returns either a
// file path to open or a hostname extracted from a kpass://fill?host= URI.
func parseInitialArg() (filePath, fillHost string) {
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "-") {
			continue
		}
		if strings.HasPrefix(arg, "kpass://") {
			u, err := url.Parse(arg)
			if err == nil && u.Host == "fill" {
				fillHost = u.Query().Get("host")
			}
			return "", fillHost
		}
		return arg, ""
	}
	return "", ""
}
