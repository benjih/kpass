import QtQuick
import QtQuick.Controls
import QtQuick.Layouts

Item {
    id: entriesPage

    property string databasePath
    property string currentGroup: ""
    property var entries: []

    signal entrySelected(int index, var entryData)
    signal addEntryRequested()
    signal saveDatabaseRequested()
    signal closeDatabaseRequested()
    signal notify(string message)

    function toast(message) {
        notify(message)
    }

    ColumnLayout {
        anchors.fill: parent
        spacing: 0

        RowLayout {
            Layout.fillWidth: true
            Layout.margins: 8
            spacing: 8

            TextField {
                id: searchField
                Layout.fillWidth: true
                placeholderText: "Search entries..."
            }

            ToolButton {
                text: "Save"
                onClicked: entriesPage.saveDatabaseRequested()
            }

            ToolButton {
                text: "Close"
                onClicked: entriesPage.closeDatabaseRequested()
            }
        }

        Label {
            text: entriesPage.visibleCount + " items found"
            Layout.leftMargin: 12
            Layout.bottomMargin: 4
            visible: entriesPage.visibleCount > 0
        }

        ListView {
            id: listView
            Layout.fillWidth: true
            Layout.fillHeight: true
            clip: true
            model: entriesPage.entries

            delegate: ItemDelegate {
                id: delegate
                width: listView.width
                visible: {
                    var inGroup = entriesPage.currentGroup === ""
                            || entriesPage.currentGroup === "All Entries"
                            || modelData.group === entriesPage.currentGroup
                    var q = searchField.text.toLowerCase()
                    var matchesSearch = q === ""
                            || modelData.title.toLowerCase().indexOf(q) !== -1
                            || modelData.username.toLowerCase().indexOf(q) !== -1
                    return inGroup && matchesSearch
                }
                height: visible ? implicitHeight : 0

                function openDetails() {
                    entriesPage.entrySelected(index, {
                        title: modelData.title,
                        username: modelData.username,
                        password: modelData.password,
                        url: modelData.url,
                        notes: modelData.notes
                    })
                }

                contentItem: RowLayout {
                    spacing: 12
                    implicitHeight: 48

                    ColumnLayout {
                        Layout.fillWidth: true
                        spacing: 0
                        Label {
                            text: modelData.title
                            elide: Text.ElideRight
                            font.bold: true
                            Layout.fillWidth: true
                        }
                        Label {
                            text: modelData.username
                            elide: Text.ElideRight
                            color: palette.placeholderText
                            Layout.fillWidth: true
                        }
                    }

                    ToolButton {
                        text: "User"
                        onClicked: {
                            databaseManager.copyToClipboard(modelData.username)
                            entriesPage.toast("Username copied")
                        }
                    }

                    ToolButton {
                        text: "Pass"
                        onClicked: {
                            databaseManager.copyToClipboard(modelData.password)
                            entriesPage.toast("Password copied")
                        }
                    }
                }

                onClicked: {
                    listView.currentIndex = index
                    openDetails()
                }
            }
        }
    }

    readonly property int visibleCount: {
        var count = 0
        for (var i = 0; i < entries.length; ++i) {
            var entry = entries[i]
            var inGroup = currentGroup === "" || currentGroup === "All Entries"
                    || entry.group === currentGroup
            var q = searchField.text.toLowerCase()
            var matchesSearch = q === ""
                    || entry.title.toLowerCase().indexOf(q) !== -1
                    || entry.username.toLowerCase().indexOf(q) !== -1
            if (inGroup && matchesSearch) {
                count++
            }
        }
        return count
    }
}
