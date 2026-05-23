#ifndef DATABASEMANAGER_H
#define DATABASEMANAGER_H

#include <QObject>
#include <QString>
#include <QVariantList>
#include <QStringList>

class DatabaseManager : public QObject
{
    Q_OBJECT
    Q_PROPERTY(QVariantList entries READ entries NOTIFY entriesChanged)
    Q_PROPERTY(QStringList groups READ groups NOTIFY groupsChanged)
    Q_PROPERTY(QString lastError READ lastError NOTIFY lastErrorChanged)

public:
    explicit DatabaseManager(QObject *parent = nullptr);

    Q_INVOKABLE bool openDatabase(const QString &path, const QString &password);
    Q_INVOKABLE void closeDatabase();
    Q_INVOKABLE QVariantList entries() const;
    Q_INVOKABLE QStringList groups() const;
    Q_INVOKABLE void deleteEntry(int index);
    Q_INVOKABLE void updateEntry(int index, const QString &title, const QString &username,
                                 const QString &password, const QString &url, const QString &notes);
    Q_INVOKABLE static void copyToClipboard(const QString &text);
    Q_INVOKABLE void saveDatabase();
    Q_INVOKABLE QString lastError() const;

signals:
    void entriesChanged();
    void groupsChanged();
    void lastErrorChanged();

private:
    void refreshFromGo();

    QVariantList m_entries;
    QStringList m_groups;
    QString m_lastError;
};

#endif
