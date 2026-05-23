package main

import (
	"os"
	"path/filepath"

	"github.com/benjih/kpass/bridge"
	qt "github.com/mappu/miqt/qt6"
	"github.com/mappu/miqt/qt6/qml"
)

func main() {
	qt.NewQApplication(os.Args)
	qt.QCoreApplication_SetApplicationName("KPass")
	qt.QCoreApplication_SetOrganizationName("KPass")

	engine := qml.NewQQmlApplicationEngine()
	engine.RootContext().SetContextProperty("databaseManager", bridge.NewDatabaseManager())

	qmlDir := filepath.Join(projectRoot(), "qml")
	mainQML := filepath.Join(qmlDir, "Main.qml")
	engine.AddImportPath(qmlDir)
	engine.Load(qt.QUrl_FromLocalFile(mainQML))

	if len(engine.RootObjects()) == 0 {
		os.Exit(1)
	}

	qt.QApplication_Exec()
}

func projectRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}
