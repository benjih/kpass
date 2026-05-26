package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/benjih/kpass/bridge"
	qt "github.com/mappu/miqt/qt6"
	"github.com/mappu/miqt/qt6/qml"
)

func init() {
	// Qt reads QT_PLUGIN_PATH and XDG_DATA_DIRS before QApplication starts.
	setupRuntimeEnvironment()
}

func main() {
	qt.NewQApplication(os.Args)
	setupKDEAppearance()

	engine := qml.NewQQmlApplicationEngine()
	engine.RootContext().SetContextProperty("databaseManager", bridge.NewDatabaseManager())

	root := projectRoot()
	engine.RootContext().SetContextProperty2(
		"appIconUrl",
		qt.NewQVariant11(qt.QUrl_FromLocalFile(filepath.Join(root, "assets", "KPass.png")).ToString()),
	)

	qmlDir := filepath.Join(root, "qml")
	mainQML := filepath.Join(qmlDir, "Main.qml")
	engine.AddImportPath(qmlDir)
	for _, importPath := range strings.Split(os.Getenv("QML2_IMPORT_PATH"), ":") {
		if importPath != "" {
			engine.AddImportPath(importPath)
		}
	}
	engine.Load(qt.QUrl_FromLocalFile(mainQML))

	if len(engine.RootObjects()) == 0 {
		os.Exit(1)
	}

	qt.QApplication_Exec()
}
