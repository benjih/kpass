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

    signal openFileRequested()
    signal recentFileSelected(string url)
    signal databaseUnlocked(string url, string password)

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
}
