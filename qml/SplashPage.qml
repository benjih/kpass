import QtQuick
import QtCore
import QtQuick.Controls as Controls
import QtQuick.Layouts
import QtQuick.Dialogs
import org.kde.kirigami as Kirigami
import Tr

Kirigami.Page {
    id: splashPage

    title: ""
    globalToolBarStyle: Kirigami.ApplicationHeaderStyle.None

    property var recentFiles: []
    property string initialFilePath: ""

    signal openFileRequested()
    signal recentFileSelected(string url)
    signal databaseUnlocked(string url, string password)
    signal databaseCreated(string url, string password)

    Component.onCompleted: {
        if (initialFilePath !== "") {
            passwordDialog.databaseUrl = initialFilePath
            passwordDialog.open()
        }
    }

    ColumnLayout {
        width: parent.width
        height: parent.height
        spacing: Kirigami.Units.largeSpacing * 2

        Item {
            Layout.fillHeight: true
        }

        Item {
            Layout.fillWidth: true
            Layout.preferredHeight: Kirigami.Units.iconSizes.huge

            Image {
                anchors.centerIn: parent
                width: Kirigami.Units.iconSizes.huge
                height: Kirigami.Units.iconSizes.huge
                source: typeof appIconUrl !== "undefined" && appIconUrl !== ""
                    ? appIconUrl
                    : Qt.resolvedUrl("../assets/KPass.png")
                fillMode: Image.PreserveAspectFit
                smooth: true
            }
        }

        ColumnLayout {
            spacing: Kirigami.Units.smallSpacing
            Layout.fillWidth: true
            Layout.alignment: Qt.AlignHCenter

            Kirigami.Heading {
                text: Tr.i18n("KPass")
                level: 1
                Layout.alignment: Qt.AlignHCenter
            }

            Controls.Label {
                text: Tr.i18n("Securely manage your passwords")
                Layout.alignment: Qt.AlignHCenter
                color: Kirigami.Theme.disabledTextColor
            }
        }

        Controls.Button {
            text: Tr.i18n("Open Database")
            icon.name: "document-open"
            Layout.preferredWidth: Kirigami.Units.gridUnit * 12
            Layout.alignment: Qt.AlignHCenter
            highlighted: true
            onClicked: fileDialog.open()
        }

        Controls.Button {
            text: Tr.i18n("Create New Database")
            icon.name: "document-new"
            Layout.preferredWidth: Kirigami.Units.gridUnit * 12
            Layout.alignment: Qt.AlignHCenter
            onClicked: createFileDialog.open()
        }

        ColumnLayout {
            visible: splashPage.recentFiles.length > 0
            Layout.alignment: Qt.AlignHCenter
            Layout.topMargin: Kirigami.Units.largeSpacing
            spacing: Kirigami.Units.smallSpacing

            Kirigami.Heading {
                text: Tr.i18n("Recent Files")
                level: 4
                Layout.alignment: Qt.AlignHCenter
            }

            Repeater {
                model: splashPage.recentFiles
                delegate: Controls.Button {
                    text: {
                        var parts = modelData.split('/')
                        return parts[parts.length - 1]
                    }
                    icon.name: "document-open-recent"
                    flat: true
                    onClicked: {
                        passwordDialog.databaseUrl = modelData
                        passwordDialog.open()
                    }
                    Layout.alignment: Qt.AlignHCenter

                    Controls.ToolTip.visible: hovered
                    Controls.ToolTip.text: modelData
                }
            }
        }

        Item {
            Layout.fillHeight: true
        }
    }

    FileDialog {
        id: fileDialog
        title: Tr.i18n("Open Password Database")
        currentFolder: StandardPaths.standardLocations(StandardPaths.HomeLocation)[0]
        nameFilters: ["KeePass Database (*.kdbx)", "All files (*)"]
        onAccepted: {
            passwordDialog.databaseUrl = selectedFile
            passwordDialog.open()
        }
    }

    Controls.Dialog {
        id: passwordDialog
        title: Tr.i18n("Unlock Database")
        anchors.centerIn: parent
        modal: true
        width: Kirigami.Units.gridUnit * 20
        standardButtons: Controls.Dialog.Ok | Controls.Dialog.Cancel

        property string databaseUrl: ""

        onOpened: passwordField.forceActiveFocus()

        onAccepted: {
            splashPage.databaseUnlocked(databaseUrl, passwordField.text)
            passwordField.text = ""
        }

        onRejected: {
            passwordField.text = ""
        }

        contentItem: ColumnLayout {
            spacing: Kirigami.Units.smallSpacing

            Controls.Label {
                text: Tr.i18n("Enter master password:")
                font.bold: true
            }

            Controls.TextField {
                id: passwordField
                echoMode: TextInput.Password
                Layout.fillWidth: true
                placeholderText: Tr.i18n("Password")
                onAccepted: passwordDialog.accept()
            }
        }
    }

    FileDialog {
        id: createFileDialog
        title: Tr.i18n("Create Password Database")
        currentFolder: StandardPaths.standardLocations(StandardPaths.HomeLocation)[0]
        fileMode: FileDialog.SaveFile
        defaultSuffix: "kdbx"
        nameFilters: ["KeePass Database (*.kdbx)", "All files (*)"]
        onAccepted: {
            createPasswordDialog.databaseUrl = selectedFile
            createPasswordDialog.open()
        }
    }

    Controls.Dialog {
        id: createPasswordDialog
        title: Tr.i18n("Set Master Password")
        anchors.centerIn: parent
        modal: true
        width: Kirigami.Units.gridUnit * 20
        standardButtons: Controls.Dialog.Ok | Controls.Dialog.Cancel

        property string databaseUrl: ""

        onOpened: {
            createNewPasswordField.text = ""
            createConfirmField.text = ""
            createErrorLabel.text = ""
            createNewPasswordField.forceActiveFocus()
        }

        onAccepted: {
            if (createNewPasswordField.text === "") {
                createErrorLabel.text = Tr.i18n("Password cannot be empty.")
                Qt.callLater(createPasswordDialog.open)
                return
            }
            if (createNewPasswordField.text !== createConfirmField.text) {
                createErrorLabel.text = Tr.i18n("Passwords do not match.")
                Qt.callLater(createPasswordDialog.open)
                return
            }
            splashPage.databaseCreated(databaseUrl, createNewPasswordField.text)
            createNewPasswordField.text = ""
            createConfirmField.text = ""
            createErrorLabel.text = ""
        }

        onRejected: {
            createNewPasswordField.text = ""
            createConfirmField.text = ""
            createErrorLabel.text = ""
        }

        contentItem: ColumnLayout {
            spacing: Kirigami.Units.smallSpacing

            Controls.Label {
                text: Tr.i18n("Set a master password for the new database:")
                font.bold: true
                wrapMode: Text.WordWrap
                Layout.fillWidth: true
            }

            Controls.TextField {
                id: createNewPasswordField
                echoMode: TextInput.Password
                Layout.fillWidth: true
                placeholderText: Tr.i18n("Password")
            }

            Controls.TextField {
                id: createConfirmField
                echoMode: TextInput.Password
                Layout.fillWidth: true
                placeholderText: Tr.i18n("Confirm Password")
                onAccepted: createPasswordDialog.accept()
            }

            Controls.Label {
                id: createErrorLabel
                text: ""
                color: Kirigami.Theme.negativeTextColor
                visible: text !== ""
                wrapMode: Text.WordWrap
                Layout.fillWidth: true
            }
        }
    }
}
