import QtQuick
import QtQuick.Controls as Controls
import QtQuick.Layouts
import org.kde.kirigami as Kirigami
import QtCore
import Tr

Kirigami.ApplicationWindow {
    id: root

    title: Tr.i18n("KPass")

    width: 900
    height: 600

    property string currentGroup: ""
    property bool databaseOpen: false

    Settings {
        id: settings
        category: "General"
        property string recentFiles: "[]"
    }

    property var parsedRecentFiles: {
        try { return JSON.parse(settings.recentFiles) } catch(e) { return [] }
    }

    function addRecentFile(url) {
        var files = parsedRecentFiles.slice()
        var index = files.indexOf(url)
        if (index !== -1) {
            files.splice(index, 1)
        }
        files.unshift(url)
        if (files.length > 5) {
            files.pop()
        }
        settings.recentFiles = JSON.stringify(files)
    }

    globalDrawer: drawer

    Kirigami.GlobalDrawer {
        id: drawer
        enabled: databaseOpen
        handleVisible: databaseOpen
        title: Tr.i18n("Filter")
        titleIcon: "view-filter"
        handleClosedIcon.name: "view-filter"
        handleOpenIcon.name: "view-filter"
        isMenu: false

        property var actionObjects: []

        function clearActions() {
            for (var j = 0; j < actionObjects.length; ++j) {
                actionObjects[j].destroy()
            }
            actionObjects = []
            drawer.actions = []
        }

        function updateActions() {
            for (var j = 0; j < actionObjects.length; ++j) {
                actionObjects[j].destroy()
            }
            actionObjects = []

            var newActions = []

            var allAction = actionComponent.createObject(drawer, {
                "text": Tr.i18n("All Entries"),
                "groupName": "All Entries"
            })
            newActions.push(allAction)
            actionObjects.push(allAction)

            for (var i = 0; i < databaseManager.groups.length; ++i) {
                var groupName = databaseManager.groups[i]
                var action = actionComponent.createObject(drawer, {
                    "text": groupName,
                    "groupName": groupName
                })
                newActions.push(action)
                actionObjects.push(action)
            }
            drawer.actions = newActions

            if (root.currentGroup === "") {
                root.currentGroup = "All Entries"
            }
        }

        Connections {
            target: root
            function onDatabaseOpenChanged() {
                if (root.databaseOpen) {
                    drawer.updateActions()
                } else {
                    drawer.clearActions()
                }
            }
        }

        Connections {
            target: databaseManager
            function onGroupsChanged() {
                if (root.databaseOpen) {
                    drawer.updateActions()
                }
            }
        }
    }

    Component {
        id: actionComponent
        Kirigami.Action {
            property string groupName
            text: groupName
            icon.name: "folder"
            checkable: true
            checked: root.currentGroup === text
            onTriggered: root.currentGroup = text
        }
    }

    pageStack.initialPage: splashPageComponent.createObject(pageStack)

    Component {
        id: splashPageComponent
        SplashPage {
            recentFiles: root.parsedRecentFiles

            onDatabaseUnlocked: function(url, password) {
                if (databaseManager.openDatabase(url, password)) {
                    root.addRecentFile(url)
                    root.databaseOpen = true
                    pageStack.replace(entriesPageComponent.createObject(pageStack, { databasePath: url }))
                } else {
                    var error = databaseManager.lastError
                    if (error === "") {
                        error = Tr.i18n("Failed to open database. Please check your password.")
                    }
                    showPassiveNotification(Tr.i18n("Error: %1", error))
                }
            }
        }
    }

    Component {
        id: entriesPageComponent
        EntriesPage {
            currentGroup: root.currentGroup
            entries: databaseManager.entries

            onEntrySelected: function(index, entryData) {
                pageStack.push(detailsPageComponent.createObject(pageStack, {
                    entryIndex: index,
                    entryTitle: entryData.title,
                    entryUsername: entryData.username,
                    entryPassword: entryData.password,
                    entryUrl: entryData.url,
                    entryNotes: entryData.notes,
                    entryIcon: entryData.icon
                }))
            }

            onAddEntryRequested: {
                showPassiveNotification(Tr.i18n("Add new entry dialog would open here"))
            }

            onSaveDatabaseRequested: {
                databaseManager.saveDatabase()
                if (databaseManager.lastError !== "") {
                    showPassiveNotification(Tr.i18n("Error saving: %1", databaseManager.lastError))
                } else {
                    showPassiveNotification(Tr.i18n("Database saved"))
                }
            }

            onCloseDatabaseRequested: {
                databaseManager.closeDatabase()
                root.currentGroup = ""
                root.databaseOpen = false
                pageStack.replace(splashPageComponent.createObject(pageStack))
                showPassiveNotification(Tr.i18n("Database closed"))
            }
        }
    }

    Component {
        id: detailsPageComponent
        DetailsPage {
            onEntryUpdated: function(index, title, username, password, url, notes) {
                databaseManager.updateEntry(index, title, username, password, url, notes)
            }

            onEntryDeleted: function(index) {
                databaseManager.deleteEntry(index)
                pageStack.pop()
                showPassiveNotification(Tr.i18n("Entry deleted"))
            }

            onGoBackRequested: {
                pageStack.pop()
            }
        }
    }
}
