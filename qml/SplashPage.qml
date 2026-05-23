import QtQuick
import QtQuick.Controls
import QtQuick.Layouts
import QtQuick.Dialogs

Item {
    id: splashPage

    property var recentFiles: []

    signal databaseUnlocked(string url, string password)

    ColumnLayout {
        anchors.centerIn: parent
        spacing: 24

        Label {
            text: "KPass"
            font.pixelSize: 32
            font.bold: true
            Layout.alignment: Qt.AlignHCenter
        }

        Label {
            text: "Securely manage your passwords"
            color: palette.placeholderText
            Layout.alignment: Qt.AlignHCenter
        }

        Button {
            text: "Open Database"
            Layout.preferredWidth: 240
            Layout.alignment: Qt.AlignHCenter
            onClicked: fileDialog.open()
        }

        ColumnLayout {
            visible: splashPage.recentFiles.length > 0
            Layout.alignment: Qt.AlignHCenter
            Layout.topMargin: 16
            spacing: 8

            Label {
                text: "Recent Files"
                font.bold: true
                Layout.alignment: Qt.AlignHCenter
            }

            Repeater {
                model: splashPage.recentFiles
                delegate: Button {
                    text: {
                        var parts = modelData.split('/')
                        return parts[parts.length - 1]
                    }
                    flat: true
                    onClicked: {
                        passwordDialog.databaseUrl = modelData
                        passwordDialog.open()
                    }
                    Layout.alignment: Qt.AlignHCenter

                    ToolTip.visible: hovered
                    ToolTip.text: modelData
                }
            }
        }
    }

    FileDialog {
        id: fileDialog
        title: "Open Password Database"
        nameFilters: ["KeePass Database (*.kdbx)", "All files (*)"]
        onAccepted: {
            passwordDialog.databaseUrl = selectedFile
            passwordDialog.open()
        }
    }

    Dialog {
        id: passwordDialog
        title: "Unlock Database"
        anchors.centerIn: parent
        modal: true
        standardButtons: Dialog.Ok | Dialog.Cancel

        property string databaseUrl: ""

        onOpened: passwordField.forceActiveFocus()

        onAccepted: {
            splashPage.databaseUnlocked(databaseUrl, passwordField.text)
            passwordField.text = ""
        }

        onRejected: passwordField.text = ""

        ColumnLayout {
            spacing: 8

            Label {
                text: "Enter master password:"
                font.bold: true
            }

            TextField {
                id: passwordField
                echoMode: TextInput.Password
                Layout.fillWidth: true
                placeholderText: "Password"
                onAccepted: passwordDialog.accept()
            }
        }
    }
}
