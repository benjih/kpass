import QtQuick
import QtQuick.Controls as Controls
import QtQuick.Layouts
import org.kde.kirigami as Kirigami
import org.kde.kirigami.primitives as KirigamiPrimitives
import Tr

Kirigami.ScrollablePage {
    id: entriesPage

    property string databasePath
    property string currentGroup: ""
    property var entries: []

    signal entrySelected(int index, var entryData)
    signal addEntryRequested()
    signal saveDatabaseRequested()
    signal closeDatabaseRequested()

    title: currentGroup

    // Primary action stays in the toolbar; others overflow (like KF5 actions.main / contextualActions).
    actions: [
        Kirigami.Action {
            icon.name: "list-add"
            text: Tr.i18n("Add Entry")
            displayHint: Kirigami.DisplayHint.KeepVisible
            onTriggered: entriesPage.addEntryRequested()
        },
        Kirigami.Action {
            icon.name: "search"
            text: Tr.i18n("Search")
            shortcut: StandardKey.Find
            onTriggered: searchFocusTimer.restart()
        },
        Kirigami.Action {
            icon.name: "document-save"
            text: Tr.i18n("Save Database")
            shortcut: StandardKey.Save
            onTriggered: entriesPage.saveDatabaseRequested()
        },
        Kirigami.Action {
            icon.name: "window-close"
            text: Tr.i18n("Close Database")
            onTriggered: entriesPage.closeDatabaseRequested()
        }
    ]

    property int visibleCount: {
        var count = 0
        for (var i = 0; i < entries.length; ++i) {
            var entry = entries[i]
            var inGroup = currentGroup === "" || currentGroup === "All Entries" || entry["group"] === currentGroup
            var matchesSearch = searchField.text === ""
                    || entry.title.toLowerCase().indexOf(searchField.text.toLowerCase()) !== -1
                    || entry.username.toLowerCase().indexOf(searchField.text.toLowerCase()) !== -1
            if (inGroup && matchesSearch) {
                count++
            }
        }
        return count
    }

    Timer {
        id: searchFocusTimer
        interval: 16
        repeat: true
        property int attempts: 0
        onTriggered: {
            searchField.forceActiveFocus(Qt.ShortcutFocusReason)
            if (searchField.activeFocus || ++attempts >= 10) {
                stop()
                attempts = 0
            }
        }
    }

    header: ColumnLayout {
        spacing: 0
        Kirigami.SearchField {
            id: searchField
            Layout.fillWidth: true
            Layout.leftMargin: Kirigami.Units.smallSpacing
            Layout.rightMargin: Kirigami.Units.smallSpacing
            Layout.topMargin: Kirigami.Units.smallSpacing
            placeholderText: Tr.i18n("Search entries...")
        }
    }

    footer: Kirigami.Heading {
        text: Tr.i18n("%1 items found", entriesPage.visibleCount)
        level: 3
        visible: entriesPage.visibleCount > 0
        width: parent.width
        padding: Kirigami.Units.largeSpacing
    }

    ListView {
        id: listView
        model: entriesPage.entries
        delegate: Kirigami.SwipeListItem {
            id: delegate
            visible: {
                var inGroup = entriesPage.currentGroup === "" || entriesPage.currentGroup === "All Entries" || modelData["group"] === entriesPage.currentGroup
                var matchesSearch = searchField.text === ""
                        || modelData.title.toLowerCase().indexOf(searchField.text.toLowerCase()) !== -1
                        || modelData.username.toLowerCase().indexOf(searchField.text.toLowerCase()) !== -1
                return inGroup && matchesSearch
            }
            height: visible ? implicitHeight : 0
            highlighted: ListView.isCurrentItem

            function openDetails() {
                entriesPage.entrySelected(index, {
                    title: modelData.title,
                    username: modelData.username,
                    password: modelData.password,
                    url: modelData.url,
                    notes: modelData.notes,
                    icon: "key-enter"
                })
            }

            contentItem: RowLayout {
                spacing: Kirigami.Units.smallSpacing
                implicitHeight: Kirigami.Units.gridUnit * 3

                KirigamiPrimitives.Icon {
                    source: "key-enter"
                    Layout.preferredWidth: Kirigami.Units.iconSizes.small
                    Layout.preferredHeight: Kirigami.Units.iconSizes.small
                    Layout.leftMargin: Kirigami.Units.largeSpacing
                }

                ColumnLayout {
                    Layout.fillWidth: true
                    spacing: 0
                    Controls.Label {
                        text: modelData.title
                        elide: Text.ElideRight
                        font.bold: true
                        Layout.fillWidth: true
                    }
                    Controls.Label {
                        text: modelData.username
                        elide: Text.ElideRight
                        color: Kirigami.Theme.disabledTextColor
                        Layout.fillWidth: true
                    }
                }

                Item {
                    Layout.preferredWidth: Kirigami.Units.largeSpacing
                }
            }

            actions: [
                Kirigami.Action {
                    icon.name: "user-identity"
                    tooltip: Tr.i18n("Copy Username")
                    onTriggered: {
                        databaseManager.copyToClipboard(modelData.username)
                        showPassiveNotification(Tr.i18n("Username copied"))
                    }
                },
                Kirigami.Action {
                    icon.name: "edit-copy"
                    tooltip: Tr.i18n("Copy Password")
                    onTriggered: {
                        databaseManager.copyToClipboard(modelData.password)
                        showPassiveNotification(Tr.i18n("Password copied"))
                    }
                }
            ]
            onClicked: {
                listView.currentIndex = index
                openDetails()
            }
        }
    }
}
