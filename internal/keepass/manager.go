package keepass

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/tobischo/gokeepasslib/v3"
)

// Manager holds an opened KeePass database and a flat entry index for QML.
type Manager struct {
	db         *gokeepasslib.Database
	filePath   string
	entries    []Entry
	groups     []string
	lastError  string
	isModified bool
}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) LastError() string {
	return m.lastError
}

func (m *Manager) Entries() []Entry {
	return m.entries
}

func (m *Manager) Groups() []string {
	return m.groups
}

func (m *Manager) IsOpen() bool {
	return m.db != nil
}

func (m *Manager) Open(path, password string) bool {
	m.lastError = ""
	path = fileURLToPath(path)

	file, err := os.Open(path)
	if err != nil {
		m.lastError = err.Error()
		return false
	}
	defer file.Close()

	db := gokeepasslib.NewDatabase()
	db.Credentials = gokeepasslib.NewPasswordCredentials(password)
	if err := gokeepasslib.NewDecoder(file).Decode(db); err != nil {
		m.lastError = err.Error()
		return false
	}

	if err := db.UnlockProtectedEntries(); err != nil {
		m.lastError = err.Error()
		return false
	}
	// Entries remain unlocked for the session; the encoder re-locks on save.

	m.db = db
	m.filePath = path
	m.rebuildIndex()
	m.isModified = false
	return true
}

func (m *Manager) Close() {
	m.db = nil
	m.filePath = ""
	m.entries = nil
	m.groups = nil
	m.lastError = ""
	m.isModified = false
}

func (m *Manager) Save() error {
	if m.db == nil || m.filePath == "" {
		err := fmt.Errorf("no database open")
		m.lastError = err.Error()
		return err
	}

	// Write to a temp file in the same directory so the final os.Rename is
	// atomic (same filesystem). If anything fails before the rename, the
	// original .kdbx file is untouched.
	dir := filepath.Dir(m.filePath)
	tmp, err := os.CreateTemp(dir, ".kpass-save-*")
	if err != nil {
		m.lastError = err.Error()
		return err
	}
	tmpName := tmp.Name()

	// Capture both errors before removing the temp file so we don't lose
	// encodeErr if Close also fails.
	encodeErr := gokeepasslib.NewEncoder(tmp).Encode(m.db)
	closeErr := tmp.Close()
	if encodeErr != nil || closeErr != nil {
		os.Remove(tmpName)
		if encodeErr != nil {
			m.lastError = encodeErr.Error()
			return encodeErr
		}
		m.lastError = closeErr.Error()
		return closeErr
	}

	// Atomic swap — readers see either the old file or the new file, never a
	// partial write.
	if err := os.Rename(tmpName, m.filePath); err != nil {
		os.Remove(tmpName)
		m.lastError = err.Error()
		return err
	}

	m.lastError = ""
	m.isModified = false
	return nil
}

func (m *Manager) DeleteEntry(index int) bool {
	if m.db == nil || index < 0 || index >= len(m.entries) {
		return false
	}
	uuid := m.entries[index].UUID
	if !m.deleteByUUID(uuid) {
		return false
	}
	m.rebuildIndex()
	m.isModified = true
	return true
}

func (m *Manager) UpdateEntry(index int, title, username, password, url, notes string) bool {
	if m.db == nil || index < 0 || index >= len(m.entries) {
		return false
	}
	uuid := m.entries[index].UUID
	updated := map[string]string{
		"Title": title,
		"User":  username,
		"URL":   url,
		"Notes": notes,
	}
	if password != "" {
		updated["Password"] = password
	}
	if !m.updateByUUID(uuid, updated) {
		return false
	}
	m.rebuildIndex()
	m.isModified = true
	return true
}

func (m *Manager) rebuildIndex() {
	m.entries = nil
	m.groups = nil
	seenGroups := map[string]struct{}{}
	if m.db == nil {
		return
	}
	for i := range m.db.Content.Root.Groups {
		m.walkGroup(&m.db.Content.Root.Groups[i], seenGroups)
	}
}

func (m *Manager) walkGroup(group *gokeepasslib.Group, seenGroups map[string]struct{}) {
	if group == nil {
		return
	}
	if group.Name != "" {
		if _, ok := seenGroups[group.Name]; !ok {
			seenGroups[group.Name] = struct{}{}
			m.groups = append(m.groups, group.Name)
		}
	}

	for _, entry := range group.Entries {
		m.entries = append(m.entries, entryFromLib(group.Name, entry))
	}
	for i := range group.Groups {
		m.walkGroup(&group.Groups[i], seenGroups)
	}
}

func entryFromLib(groupName string, entry gokeepasslib.Entry) Entry {
	return Entry{
		Group:    groupName,
		Title:    entry.GetTitle(),
		Username: valueContent(entry, "UserName", "User"),
		UUID:     uuidToString(entry.UUID),
		Password: valueContent(entry, "Password"),
		URL:      valueContent(entry, "URL"),
		Notes:    valueContent(entry, "Notes"),
	}
}

func valueContent(entry gokeepasslib.Entry, keys ...string) string {
	for _, key := range keys {
		if v := entry.Get(key); v != nil {
			return v.Value.Content
		}
	}
	return ""
}

func uuidToString(id [16]byte) string {
	return fmt.Sprintf("%x", id)
}

func (m *Manager) deleteByUUID(id string) bool {
	for i := range m.db.Content.Root.Groups {
		if m.deleteInGroup(id, &m.db.Content.Root.Groups[i]) {
			return true
		}
	}
	return false
}

func (m *Manager) deleteInGroup(id string, group *gokeepasslib.Group) bool {
	for i, entry := range group.Entries {
		if uuidToString(entry.UUID) == id {
			group.Entries = append(group.Entries[:i], group.Entries[i+1:]...)
			return true
		}
	}
	for i := range group.Groups {
		if m.deleteInGroup(id, &group.Groups[i]) {
			return true
		}
	}
	return false
}

func (m *Manager) updateByUUID(id string, data map[string]string) bool {
	for i := range m.db.Content.Root.Groups {
		if m.updateInGroup(id, &m.db.Content.Root.Groups[i], data) {
			return true
		}
	}
	return false
}

func (m *Manager) updateInGroup(id string, group *gokeepasslib.Group, data map[string]string) bool {
	for i := range group.Entries {
		entry := &group.Entries[i]
		if uuidToString(entry.UUID) != id {
			continue
		}
		if title, ok := data["Title"]; ok {
			if v := entry.Get("Title"); v != nil {
				v.Value.Content = title
			}
		}
		for key, val := range data {
			if key == "Title" {
				continue
			}
			libKey := key
			if key == "User" {
				libKey = "UserName"
			}
			if v := entry.Get(libKey); v != nil {
				v.Value.Content = val
			} else {
				entry.Values = append(entry.Values, gokeepasslib.ValueData{
					Key:   libKey,
					Value: gokeepasslib.V{Content: val},
				})
			}
		}
		return true
	}
	for i := range group.Groups {
		if m.updateInGroup(id, &group.Groups[i], data) {
			return true
		}
	}
	return false
}

func fileURLToPath(path string) string {
	path = strings.TrimSpace(path)
	if strings.HasPrefix(path, "file://") {
		path = strings.TrimPrefix(path, "file://")
		if len(path) > 2 && path[0] == '/' && path[2] == ':' {
			path = path[1:]
		}
		if decoded, err := url.PathUnescape(path); err == nil {
			path = decoded
		}
	}
	return path
}
