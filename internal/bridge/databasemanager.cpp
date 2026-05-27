#include "databasemanager.h"

#include <QJsonArray>
#include <QJsonDocument>
#include <QJsonObject>
#include <QUrl>

extern "C" {
int goOpenDatabase(const char *path, const char *password);
void goCloseDatabase();
char *goGetEntriesJSON();
char *goGetGroupsJSON();
char *goGetLastError();
int goDeleteEntry(int index);
int goUpdateEntry(int index, const char *title, const char *username, const char *password,
                  const char *url, const char *notes);
int goSaveDatabase();
void goFreeString(char *s);
void goCopyToClipboard(const char *text);
}

// parseEntriesJson deserialises the JSON produced by goGetEntriesJSON() into a
// QVariantList of QVariantMaps. It takes ownership of json and frees it via
// goFreeString immediately after parsing, before any early return.
static QVariantList parseEntriesJson(const char *json)
{
    QVariantList out;
    if (!json || !*json) {
        return out;
    }
    const QJsonDocument doc = QJsonDocument::fromJson(QByteArray(json));
    goFreeString(const_cast<char *>(json));
    if (!doc.isArray()) {
        return out;
    }
    const QJsonArray arr = doc.array();
    for (const QJsonValue &val : arr) {
        const QJsonObject obj = val.toObject();
        QVariantMap map;
        map.insert(QStringLiteral("group"), obj.value(QStringLiteral("group")).toString());
        map.insert(QStringLiteral("title"), obj.value(QStringLiteral("title")).toString());
        map.insert(QStringLiteral("username"), obj.value(QStringLiteral("username")).toString());
        map.insert(QStringLiteral("uuid"), obj.value(QStringLiteral("uuid")).toString());
        map.insert(QStringLiteral("password"), obj.value(QStringLiteral("password")).toString());
        map.insert(QStringLiteral("url"), obj.value(QStringLiteral("url")).toString());
        map.insert(QStringLiteral("notes"), obj.value(QStringLiteral("notes")).toString());
        out.append(map);
    }
    return out;
}

// parseGroupsJson deserialises the JSON produced by goGetGroupsJSON() into a
// QStringList. Owns and frees json via goFreeString after parsing.
static QStringList parseGroupsJson(const char *json)
{
    QStringList out;
    if (!json || !*json) {
        return out;
    }
    const QJsonDocument doc = QJsonDocument::fromJson(QByteArray(json));
    goFreeString(const_cast<char *>(json));
    if (!doc.isArray()) {
        return out;
    }
    const QJsonArray arr = doc.array();
    for (const QJsonValue &val : arr) {
        out.append(val.toString());
    }
    return out;
}

DatabaseManager::DatabaseManager(QObject *parent)
    : QObject(parent)
{
}

// refreshFromGo re-fetches all mutable state from the Go side and emits the
// three change signals so QML bindings update automatically. It is the single
// sync point called after every operation that may alter entries, groups, or
// error state.
void DatabaseManager::refreshFromGo()
{
    m_entries = parseEntriesJson(goGetEntriesJSON());
    m_groups = parseGroupsJson(goGetGroupsJSON());

    char *err = goGetLastError();
    m_lastError = err ? QString::fromUtf8(err) : QString();
    goFreeString(err);

    emit entriesChanged();
    emit groupsChanged();
    emit lastErrorChanged();
}

bool DatabaseManager::openDatabase(const QString &path, const QString &password)
{
    // QML's FileDialog hands us a file:// URL; convert it to a plain path.
    // If path is already a plain path, toLocalFile() returns empty and we
    // fall back to the original string.
    QString localPath = QUrl(path).toLocalFile();
    if (localPath.isEmpty()) {
        localPath = path;
    }

    const bool ok = goOpenDatabase(localPath.toUtf8().constData(), password.toUtf8().constData()) != 0;
    refreshFromGo();
    return ok;
}

void DatabaseManager::closeDatabase()
{
    goCloseDatabase();
    refreshFromGo();
}

QVariantList DatabaseManager::entries() const
{
    return m_entries;
}

QStringList DatabaseManager::groups() const
{
    return m_groups;
}

void DatabaseManager::deleteEntry(int index)
{
    if (goDeleteEntry(index)) {
        refreshFromGo();
    }
}

void DatabaseManager::updateEntry(int index, const QString &title, const QString &username,
                                const QString &password, const QString &url, const QString &notes)
{
    if (goUpdateEntry(index, title.toUtf8().constData(), username.toUtf8().constData(),
                      password.toUtf8().constData(), url.toUtf8().constData(),
                      notes.toUtf8().constData())) {
        refreshFromGo();
    }
}

void DatabaseManager::copyToClipboard(const QString &text)
{
    goCopyToClipboard(text.toUtf8().constData());
}

void DatabaseManager::saveDatabase()
{
    goSaveDatabase();
    refreshFromGo();
}

QString DatabaseManager::lastError() const
{
    return m_lastError;
}

// newDatabaseManager is the C factory called from Go's bridge.NewDatabaseManager.
// Returning void* avoids exposing the C++ type across the cgo boundary.
extern "C" void *newDatabaseManager(void *parent)
{
    return new DatabaseManager(static_cast<QObject *>(parent));
}
