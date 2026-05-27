import QtQuick
import QtQuick.Controls as Controls
import QtQuick.Dialogs
import QtQuick.Layouts
import org.kde.kirigami as Kirigami
import org.kde.kirigami.layouts as KirigamiLayouts
import org.kde.kirigami.primitives as KirigamiPrimitives
import Tr

Kirigami.ScrollablePage {
    id: detailsPage

    leftPadding: Kirigami.Units.gridUnit * 2
    rightPadding: Kirigami.Units.gridUnit * 2
    topPadding: Kirigami.Units.gridUnit * 2
    bottomPadding: Kirigami.Units.gridUnit * 2

    property int entryIndex: -1
    property string entryTitle
    property string entryUsername
    property string entryPassword
    property string entryUrl
    property string entryNotes
    property string entryIcon
    property var entryAttachments: []
    property bool isEditing: false

    property string _pendingDownload: ""

    signal entryUpdated(int entryIdx, string newTitle, string newUsername, string newPassword, string newUrl, string newNotes)
    signal entryDeleted(int entryIdx)
    signal goBackRequested()

    title: entryTitle

    actions: [
        Kirigami.Action {
            text: detailsPage.isEditing ? Tr.i18n("Save") : Tr.i18n("Edit")
            icon.name: detailsPage.isEditing ? "document-save" : "edit-entry"
            displayHint: Kirigami.DisplayHint.KeepVisible
            onTriggered: {
                if (detailsPage.isEditing) {
                    detailsPage.entryUpdated(
                        detailsPage.entryIndex,
                        titleField.text,
                        usernameField.text,
                        passwordFieldInDetails.text,
                        urlField.text,
                        notesField.text
                    )
                    showPassiveNotification(Tr.i18n("Entry updated in memory"))
                }
                detailsPage.isEditing = !detailsPage.isEditing
            }
        },
        Kirigami.Action {
            text: Tr.i18n("Cancel")
            icon.name: "dialog-cancel"
            visible: detailsPage.isEditing
            onTriggered: {
                detailsPage.isEditing = false
                titleField.text = detailsPage.entryTitle
                usernameField.text = detailsPage.entryUsername
                passwordFieldInDetails.text = detailsPage.entryPassword
                urlField.text = detailsPage.entryUrl
                notesField.text = detailsPage.entryNotes
            }
        },
        Kirigami.Action {
            text: Tr.i18n("Delete")
            icon.name: "edit-delete"
            onTriggered: deleteDialog.open()
        },
        Kirigami.Action {
            text: Tr.i18n("Back")
            icon.name: "go-previous"
            onTriggered: detailsPage.goBackRequested()
        }
    ]

    Controls.Dialog {
        id: deleteDialog
        parent: Controls.Overlay.overlay
        title: Tr.i18n("Delete Entry")
        modal: true
        width: Kirigami.Units.gridUnit * 24
        standardButtons: Controls.Dialog.Yes | Controls.Dialog.No

        contentItem: Controls.Label {
            text: Tr.i18n("Are you sure you want to delete '%1'?", detailsPage.entryTitle)
            wrapMode: Text.WordWrap
            width: parent.width
        }

        onAccepted: {
            detailsPage.entryDeleted(detailsPage.entryIndex)
        }
    }

    KirigamiLayouts.FormLayout {
        id: formLayout
        width: parent.width

        KirigamiPrimitives.Separator {
            KirigamiLayouts.FormData.isSection: true
            KirigamiLayouts.FormData.label: Tr.i18n("Credentials")
        }

        Controls.TextField {
            id: titleField
            KirigamiLayouts.FormData.label: Tr.i18n("Title:")
            text: entryTitle
            readOnly: !detailsPage.isEditing
            Layout.fillWidth: true
        }

        RowLayout {
            KirigamiLayouts.FormData.label: Tr.i18n("Username:")
            Layout.fillWidth: true
            spacing: Kirigami.Units.smallSpacing

            Controls.TextField {
                id: usernameField
                text: entryUsername
                readOnly: !detailsPage.isEditing
                Layout.fillWidth: true
            }

            Controls.ToolButton {
                icon.name: "edit-copy"
                onClicked: {
                    databaseManager.copyToClipboard(entryUsername)
                    showPassiveNotification(Tr.i18n("Username copied"))
                }
            }
        }

        RowLayout {
            KirigamiLayouts.FormData.label: Tr.i18n("Password:")
            Layout.fillWidth: true
            spacing: Kirigami.Units.smallSpacing

            Controls.TextField {
                id: passwordFieldInDetails
                text: entryPassword
                echoMode: TextInput.Password
                readOnly: !detailsPage.isEditing
                Layout.fillWidth: true
            }

            Controls.ToolButton {
                icon.name: "edit-copy"
                onClicked: {
                    databaseManager.copyToClipboard(entryPassword)
                    showPassiveNotification(Tr.i18n("Password copied"))
                }
            }

            Controls.ToolButton {
                icon.name: passwordFieldInDetails.echoMode === TextInput.Password ? "password-show-on" : "password-hide-on"
                checkable: true
                onCheckedChanged: passwordFieldInDetails.echoMode = checked ? TextInput.Normal : TextInput.Password
            }
        }

        Controls.TextField {
            id: urlField
            KirigamiLayouts.FormData.label: Tr.i18n("URL:")
            text: entryUrl
            placeholderText: "https://example.com"
            readOnly: !detailsPage.isEditing
            Layout.fillWidth: true
        }

        KirigamiPrimitives.Separator {
            KirigamiLayouts.FormData.isSection: true
            KirigamiLayouts.FormData.label: Tr.i18n("Additional Info")
        }

        Controls.TextArea {
            id: notesField
            KirigamiLayouts.FormData.label: Tr.i18n("Notes:")
            text: entryNotes
            placeholderText: Tr.i18n("Enter notes here...")
            readOnly: !detailsPage.isEditing
            Layout.fillWidth: true
            Layout.minimumHeight: Kirigami.Units.gridUnit * 10
        }

        KirigamiPrimitives.Separator {
            KirigamiLayouts.FormData.isSection: true
            KirigamiLayouts.FormData.label: Tr.i18n("Attachments")
        }

        Repeater {
            model: detailsPage.entryAttachments
            delegate: RowLayout {
                KirigamiLayouts.FormData.label: modelData
                Layout.fillWidth: true
                spacing: Kirigami.Units.smallSpacing

                Controls.Label {
                    text: modelData
                    Layout.fillWidth: true
                    elide: Text.ElideRight
                }

                Controls.ToolButton {
                    icon.name: "document-save"
                    onClicked: {
                        detailsPage._pendingDownload = modelData
                        saveAttachmentDialog.open()
                    }
                }

                Controls.ToolButton {
                    icon.name: "edit-delete"
                    visible: detailsPage.isEditing
                    onClicked: {
                        databaseManager.deleteAttachment(detailsPage.entryIndex, modelData)
                        showPassiveNotification(Tr.i18n("Attachment removed"))
                    }
                }
            }
        }

        Controls.Label {
            KirigamiLayouts.FormData.label: ""
            text: Tr.i18n("No attachments")
            visible: detailsPage.entryAttachments.length === 0
            color: Kirigami.Theme.disabledTextColor
        }

        Controls.Button {
            KirigamiLayouts.FormData.label: ""
            visible: detailsPage.isEditing
            text: Tr.i18n("Add Attachment")
            icon.name: "mail-attachment"
            onClicked: addAttachmentDialog.open()
        }

        Item {
            KirigamiLayouts.FormData.isSection: true
        }
    }

    Connections {
        target: databaseManager
        function onEntriesChanged() {
            var ents = databaseManager.entries
            if (detailsPage.entryIndex >= 0 && detailsPage.entryIndex < ents.length) {
                detailsPage.entryAttachments = ents[detailsPage.entryIndex].attachments || []
            }
        }
    }

    FileDialog {
        id: saveAttachmentDialog
        fileMode: FileDialog.SaveFile
        title: Tr.i18n("Save Attachment")
        onAccepted: {
            databaseManager.saveAttachment(detailsPage.entryIndex, detailsPage._pendingDownload, selectedFile)
            showPassiveNotification(Tr.i18n("Attachment saved"))
        }
    }

    FileDialog {
        id: addAttachmentDialog
        fileMode: FileDialog.OpenFile
        title: Tr.i18n("Add Attachment")
        onAccepted: {
            databaseManager.addAttachment(detailsPage.entryIndex, selectedFile)
            showPassiveNotification(Tr.i18n("Attachment added"))
        }
    }
}
