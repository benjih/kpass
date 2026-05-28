#ifndef DATABASEMANAGER_H
#define DATABASEMANAGER_H

#include <QObject>
#include <QString>
#include <QVariantList>
#include <QStringList>
#include <QFileInfo>

/**
 * DatabaseManager is the QML-facing interface to the Go KeePass backend.
 *
 * miqt can wrap existing Qt types in Go but cannot declare new Q_INVOKABLE
 * methods or Q_PROPERTY bindings — those require the Qt meta-object compiler
 * (moc), which only processes C++ headers. This class exists solely to provide
 * that thin C++ shell. Every method delegates immediately to a cgo-exported Go
 * function and then calls refreshFromGo() to pull updated state back.
 *
 * QML uses this object via a context property named "db". Example:
 *   db.openDatabase(path, password)
 *   db.entries  // QVariantList, each element is a QVariantMap
 */
class DatabaseManager : public QObject
{
    Q_OBJECT

    // Live list of all entries in the open database. Each entry is a QVariantMap
    // with keys: group, title, username, uuid, password, url, notes.
    Q_PROPERTY(QVariantList entries READ entries NOTIFY entriesChanged)

    // Flat, deduplicated list of group names present in the open database.
    Q_PROPERTY(QStringList groups READ groups NOTIFY groupsChanged)

    // Human-readable description of the last error, or empty string on success.
    Q_PROPERTY(QString lastError READ lastError NOTIFY lastErrorChanged)

public:
    explicit DatabaseManager(QObject *parent = nullptr);

    // openDatabase unlocks the .kdbx file at path with password.
    // path may be a local file path or a QML file:// URL — both are handled.
    // Returns true on success; check lastError on failure.
    Q_INVOKABLE bool openDatabase(const QString &path, const QString &password);

    // closeDatabase releases the currently open database from memory.
    Q_INVOKABLE void closeDatabase();

    // entries returns the cached entry list (see Q_PROPERTY above).
    Q_INVOKABLE QVariantList entries() const;

    // groups returns the cached group list (see Q_PROPERTY above).
    Q_INVOKABLE QStringList groups() const;

    // deleteEntry removes the entry at index from the in-memory database.
    // Call saveDatabase() afterward to persist the change to disk.
    Q_INVOKABLE void deleteEntry(int index);

    // updateEntry overwrites the editable fields of the entry at index.
    // Call saveDatabase() afterward to persist the change to disk.
    Q_INVOKABLE void updateEntry(int index, const QString &title, const QString &username,
                                 const QString &password, const QString &url, const QString &notes);

    // copyToClipboard writes text to the system clipboard.
    // The Go layer handles the 10-second auto-clear timer.
    Q_INVOKABLE void copyToClipboard(const QString &text);

    // saveDatabase flushes the current in-memory state to the original .kdbx file.
    Q_INVOKABLE void saveDatabase();

    // saveAttachment reads the named attachment from the entry at index and
    // writes its bytes to savePath (a file:// URL or plain path).
    Q_INVOKABLE void saveAttachment(int index, const QString &filename, const QString &savePath);

    // addAttachment reads the file at filePath and attaches it to the entry at
    // index, using the file's base name as the attachment name.
    Q_INVOKABLE void addAttachment(int index, const QString &filePath);

    // deleteAttachment removes the named attachment from the entry at index.
    Q_INVOKABLE void deleteAttachment(int index, const QString &filename);

    // lastError returns the cached error string (see Q_PROPERTY above).
    Q_INVOKABLE QString lastError() const;

signals:
    void entriesChanged();
    void groupsChanged();
    void lastErrorChanged();

private:
    // refreshFromGo re-fetches entries, groups, and lastError from Go and
    // emits the corresponding signals. Called after every mutating operation.
    void refreshFromGo();

    QVariantList m_entries;
    QStringList m_groups;
    QString m_lastError;
};

#endif
