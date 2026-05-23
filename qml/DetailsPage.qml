import QtQuick
import QtQuick.Controls
import QtQuick.Layouts

Item {
    id: detailsPage

    property int entryIndex: -1
    property string entryTitle
    property string entryUsername
    property string entryPassword
    property string entryUrl
    property string entryNotes
    property bool isEditing: false

    signal entryUpdated(int entryIdx, string newTitle, string newUsername,
                        string newPassword, string newUrl, string newNotes)
    signal entryDeleted(int entryIdx)
    signal goBackRequested()
    signal notify(string message)

    function toast(message) {
        notify(message)
    }

    ColumnLayout {
        anchors.fill: parent
        spacing: 0

        ToolBar {
            Layout.fillWidth: true
            RowLayout {
                anchors.fill: parent
                spacing: 8

                ToolButton {
                    text: "Back"
                    onClicked: detailsPage.goBackRequested()
                }

                Label {
                    text: detailsPage.entryTitle
                    font.bold: true
                    Layout.fillWidth: true
                    elide: Text.ElideRight
                }

                ToolButton {
                    text: detailsPage.isEditing ? "Save" : "Edit"
                    onClicked: {
                        if (detailsPage.isEditing) {
                            detailsPage.entryUpdated(
                                detailsPage.entryIndex,
                                titleField.text,
                                usernameField.text,
                                passwordFieldInDetails.text,
                                urlField.text,
                                notesField.text
                            )
                            detailsPage.toast("Entry updated")
                        }
                        detailsPage.isEditing = !detailsPage.isEditing
                    }
                }

                ToolButton {
                    text: "Delete"
                    onClicked: deleteDialog.open()
                }
            }
        }

        ScrollView {
            Layout.fillWidth: true
            Layout.fillHeight: true
            Layout.margins: 16

            ColumnLayout {
                width: parent.width
                spacing: 12

                Label {
                    text: "Credentials"
                    font.bold: true
                }

                Label { text: "Title:" }
                TextField {
                    id: titleField
                    text: entryTitle
                    readOnly: !detailsPage.isEditing
                    Layout.fillWidth: true
                }

                Label { text: "Username:" }
                RowLayout {
                    Layout.fillWidth: true
                    TextField {
                        id: usernameField
                        text: entryUsername
                        readOnly: !detailsPage.isEditing
                        Layout.fillWidth: true
                    }
                    ToolButton {
                        text: "Copy"
                        onClicked: {
                            databaseManager.copyToClipboard(entryUsername)
                            detailsPage.toast("Username copied")
                        }
                    }
                }

                Label { text: "Password:" }
                RowLayout {
                    Layout.fillWidth: true
                    TextField {
                        id: passwordFieldInDetails
                        text: entryPassword
                        echoMode: TextInput.Password
                        readOnly: !detailsPage.isEditing
                        Layout.fillWidth: true
                    }
                    ToolButton {
                        text: "Copy"
                        onClicked: {
                            databaseManager.copyToClipboard(entryPassword)
                            detailsPage.toast("Password copied")
                        }
                    }
                    ToolButton {
                        text: passwordFieldInDetails.echoMode === TextInput.Password ? "Show" : "Hide"
                        onClicked: passwordFieldInDetails.echoMode =
                            passwordFieldInDetails.echoMode === TextInput.Password
                            ? TextInput.Normal : TextInput.Password
                    }
                }

                Label { text: "URL:" }
                TextField {
                    id: urlField
                    text: entryUrl
                    placeholderText: "https://example.com"
                    readOnly: !detailsPage.isEditing
                    Layout.fillWidth: true
                }

                Label {
                    text: "Notes"
                    font.bold: true
                    Layout.topMargin: 8
                }

                TextArea {
                    id: notesField
                    text: entryNotes
                    placeholderText: "Enter notes here..."
                    readOnly: !detailsPage.isEditing
                    Layout.fillWidth: true
                    Layout.minimumHeight: 120
                }
            }
        }
    }

    Dialog {
        id: deleteDialog
        title: "Delete Entry"
        anchors.centerIn: parent
        modal: true
        standardButtons: Dialog.Yes | Dialog.No

        Label {
            text: "Are you sure you want to delete '" + detailsPage.entryTitle + "'?"
        }

        onAccepted: detailsPage.entryDeleted(detailsPage.entryIndex)
    }
}
