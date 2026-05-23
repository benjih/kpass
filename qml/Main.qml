import QtQuick
import QtQuick.Controls
import QtQuick.Layouts
import Qt.labs.settings 1.0

ApplicationWindow {
    id: root

    title: "KPass"
    width: 900
    height: 600
    visible: true

    property string currentGroup: ""
    property bool databaseOpen: false

    Settings {
        id: settings
        category: "General"
        property var recentFiles: []
    }

    function addRecentFile(url) {
        var files = settings.recentFiles
        if (!Array.isArray(files)) {
            files = []
        }
        var index = files.indexOf(url)
        if (index !== -1) {
            files.splice(index, 1)
        }
        files.unshift(url)
        if (files.length > 5) {
            files.pop()
        }
        settings.recentFiles = files
    }

    function showToast(message) {
        toastLabel.text = message
        toast.open()
    }

    Drawer {
        id: groupDrawer
        edge: Qt.LeftEdge
        width: Math.min(parent.width * 0.35, 320)
        visible: false
        enabled: databaseOpen
        modal: true
        interactive: databaseOpen

        ColumnLayout {
            anchors.fill: parent
            anchors.margins: 12
            spacing: 8

            Label {
                text: "Filter"
                font.bold: true
                font.pixelSize: 18
            }

            ListView {
                id: groupList
                Layout.fillWidth: true
                Layout.fillHeight: true
                clip: true
                model: groupModel
                delegate: ItemDelegate {
                    width: groupList.width
                    text: model.name
                    highlighted: root.currentGroup === model.name
                    onClicked: {
                        root.currentGroup = model.name
                        groupDrawer.close()
                    }
                }
            }
        }
    }

    ListModel {
        id: groupModel
    }

    function rebuildGroupModel() {
        groupModel.clear()
        groupModel.append({ name: "All Entries" })
        for (var i = 0; i < databaseManager.groups.length; ++i) {
            groupModel.append({ name: databaseManager.groups[i] })
        }
        if (root.currentGroup === "") {
            root.currentGroup = "All Entries"
        }
    }

    Connections {
        target: databaseManager
        function onGroupsChanged() {
            rebuildGroupModel()
        }
    }

    header: ToolBar {
        visible: databaseOpen
        RowLayout {
            anchors.fill: parent
            anchors.margins: 4
            spacing: 8

            ToolButton {
                text: "Groups"
                onClicked: groupDrawer.open()
            }

            Label {
                Layout.fillWidth: true
                text: root.currentGroup
                elide: Text.ElideRight
                font.bold: true
            }
        }
    }

    StackView {
        id: stack
        anchors.fill: parent
        initialItem: splashPageComponent
    }

    Popup {
        id: toast
        y: root.height - implicitHeight - 24
        x: (root.width - width) / 2
        width: Math.min(root.width - 48, toastLabel.implicitWidth + 32)
        height: toastLabel.implicitHeight + 24
        modal: false
        focus: false
        closePolicy: Popup.CloseOnEscape
        padding: 12

        background: Rectangle {
            radius: 6
            color: palette.windowText
            opacity: 0.9
        }

        contentItem: Label {
            id: toastLabel
            color: palette.window
            wrapMode: Text.WordWrap
            width: parent.width
        }

        Timer {
            interval: 2500
            running: toast.opened
            onTriggered: toast.close()
        }
    }

    Component {
        id: splashPageComponent
        SplashPage {
            recentFiles: settings.recentFiles

            onDatabaseUnlocked: function(url, password) {
                if (databaseManager.openDatabase(url, password)) {
                    root.addRecentFile(url)
                    root.databaseOpen = true
                    rebuildGroupModel()
                    stack.replace(entriesPageComponent, { databasePath: url })
                } else {
                    var error = databaseManager.lastError
                    if (error === "") {
                        error = "Failed to open database. Please check your password."
                    }
                    root.showToast("Error: " + error)
                }
            }
        }
    }

    Component {
        id: entriesPageComponent
        EntriesPage {
            currentGroup: root.currentGroup
            entries: databaseManager.entries
            onNotify: function(message) { root.showToast(message) }

            onEntrySelected: function(index, entryData) {
                stack.push(detailsPageComponent, {
                    entryIndex: index,
                    entryTitle: entryData.title,
                    entryUsername: entryData.username,
                    entryPassword: entryData.password,
                    entryUrl: entryData.url,
                    entryNotes: entryData.notes
                })
            }

            onAddEntryRequested: {
                root.showToast("Add new entry is not implemented yet")
            }

            onSaveDatabaseRequested: {
                databaseManager.saveDatabase()
                root.showToast("Database saved")
            }

            onCloseDatabaseRequested: {
                databaseManager.closeDatabase()
                root.currentGroup = ""
                root.databaseOpen = false
                stack.replace(splashPageComponent)
                root.showToast("Database closed")
            }
        }
    }

    Component {
        id: detailsPageComponent
        DetailsPage {
            onNotify: function(message) { root.showToast(message) }

            onEntryUpdated: function(index, title, username, password, url, notes) {
                databaseManager.updateEntry(index, title, username, password, url, notes)
            }

            onEntryDeleted: function(index) {
                databaseManager.deleteEntry(index)
                stack.pop()
                root.showToast("Entry deleted")
            }

            onGoBackRequested: stack.pop()
        }
    }
}
