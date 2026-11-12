package keepass

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tobischo/gokeepasslib/v3"
	"github.com/tobischo/gokeepasslib/v3/wrappers"
)

// --- fileURLToPath ---

func TestFileURLToPath(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "plain path unchanged",
			input: "/home/user/db.kdbx",
			want:  "/home/user/db.kdbx",
		},
		{
			name:  "file URI stripped",
			input: "file:///home/user/db.kdbx",
			want:  "/home/user/db.kdbx",
		},
		{
			name:  "URL-encoded spaces decoded",
			input: "file:///home/user/my%20db.kdbx",
			want:  "/home/user/my db.kdbx",
		},
		{
			name:  "Windows drive letter strip",
			input: "file:///C:/Users/db.kdbx",
			want:  "C:/Users/db.kdbx",
		},
		{
			name:  "leading and trailing whitespace trimmed",
			input: "  /some/path.kdbx  ",
			want:  "/some/path.kdbx",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fileURLToPath(tt.input); got != tt.want {
				t.Errorf("fileURLToPath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// --- uuidToString ---

func TestUUIDToString(t *testing.T) {
	tests := []struct {
		name  string
		input [16]byte
		want  string
	}{
		{
			name:  "all zeros",
			input: [16]byte{},
			want:  "00000000000000000000000000000000",
		},
		{
			name:  "known bytes",
			input: [16]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10},
			want:  "0102030405060708090a0b0c0d0e0f10",
		},
		{
			name:  "max bytes",
			input: [16]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			want:  "ffffffffffffffffffffffffffffffff",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := uuidToString(tt.input); got != tt.want {
				t.Errorf("uuidToString(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// --- valueContent ---

func makeEntry(pairs ...string) gokeepasslib.Entry {
	entry := gokeepasslib.Entry{}
	for i := 0; i+1 < len(pairs); i += 2 {
		entry.Values = append(entry.Values, gokeepasslib.ValueData{
			Key:   pairs[i],
			Value: gokeepasslib.V{Content: pairs[i+1]},
		})
	}
	return entry
}

func TestValueContent(t *testing.T) {
	tests := []struct {
		name  string
		entry gokeepasslib.Entry
		keys  []string
		want  string
	}{
		{
			name:  "key present",
			entry: makeEntry("Title", "My Bank"),
			keys:  []string{"Title"},
			want:  "My Bank",
		},
		{
			name:  "first key absent fallback to second",
			entry: makeEntry("UserName", "alice"),
			keys:  []string{"User", "UserName"},
			want:  "alice",
		},
		{
			name:  "first key wins",
			entry: makeEntry("UserName", "alice", "User", "bob"),
			keys:  []string{"UserName", "User"},
			want:  "alice",
		},
		{
			name:  "no matching key returns empty",
			entry: makeEntry("Title", "My Bank"),
			keys:  []string{"Password"},
			want:  "",
		},
		{
			name:  "no keys argument returns empty",
			entry: makeEntry("Title", "My Bank"),
			keys:  nil,
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := valueContent(tt.entry, tt.keys...); got != tt.want {
				t.Errorf("valueContent(%v) = %q, want %q", tt.keys, got, tt.want)
			}
		})
	}
}

// --- entryFromLib ---

func TestEntryFromLib(t *testing.T) {
	uuid := gokeepasslib.UUID{0xde, 0xad, 0xbe, 0xef, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	libEntry := gokeepasslib.Entry{
		UUID: uuid,
		Values: []gokeepasslib.ValueData{
			{Key: "Title", Value: gokeepasslib.V{Content: "GitHub"}},
			{Key: "UserName", Value: gokeepasslib.V{Content: "alice"}},
			{Key: "Password", Value: gokeepasslib.V{Content: "s3cr3t"}},
			{Key: "URL", Value: gokeepasslib.V{Content: "https://github.com"}},
			{Key: "Notes", Value: gokeepasslib.V{Content: "work account"}},
		},
		Binaries: []gokeepasslib.BinaryReference{
			gokeepasslib.NewBinaryReference("key.pem", 0),
		},
	}

	got := entryFromLib("Dev", libEntry)

	if got.Group != "Dev" {
		t.Errorf("Group = %q, want %q", got.Group, "Dev")
	}
	if got.Title != "GitHub" {
		t.Errorf("Title = %q, want %q", got.Title, "GitHub")
	}
	if got.Username != "alice" {
		t.Errorf("Username = %q, want %q", got.Username, "alice")
	}
	if got.Password != "s3cr3t" {
		t.Errorf("Password = %q, want %q", got.Password, "s3cr3t")
	}
	if got.URL != "https://github.com" {
		t.Errorf("URL = %q, want %q", got.URL, "https://github.com")
	}
	if got.Notes != "work account" {
		t.Errorf("Notes = %q, want %q", got.Notes, "work account")
	}
	if got.UUID != uuidToString([16]byte(uuid)) {
		t.Errorf("UUID = %q, want %q", got.UUID, uuidToString([16]byte(uuid)))
	}
	if len(got.Attachments) != 1 || got.Attachments[0].Name != "key.pem" {
		t.Errorf("Attachments = %v, want [{Name:key.pem}]", got.Attachments)
	}
}

func TestEntryFromLibNoAttachments(t *testing.T) {
	libEntry := gokeepasslib.Entry{
		Values: []gokeepasslib.ValueData{
			{Key: "Title", Value: gokeepasslib.V{Content: "Empty"}},
		},
	}
	got := entryFromLib("Root", libEntry)
	if len(got.Attachments) != 0 {
		t.Errorf("expected empty Attachments, got %v", got.Attachments)
	}
}

func TestUpdateEntryPasswordPersistsAcrossSaveAndReopen(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.kdbx")
	masterPassword := "master-secret"

	db := gokeepasslib.NewDatabase()
	db.Credentials = gokeepasslib.NewPasswordCredentials(masterPassword)
	entry := gokeepasslib.NewEntry()
	entry.Values = []gokeepasslib.ValueData{
		{Key: "Title", Value: gokeepasslib.V{Content: "Email"}},
		{Key: "UserName", Value: gokeepasslib.V{Content: "alice@example.com"}},
		{Key: "Password", Value: gokeepasslib.V{Content: "initial-pass!", Protected: wrappers.NewBoolWrapper(true)}},
		{Key: "URL", Value: gokeepasslib.V{Content: "https://mail.example.com"}},
		{Key: "Notes", Value: gokeepasslib.V{Content: "initial note"}},
	}
	group := gokeepasslib.NewGroup()
	group.Name = "Personal"
	group.Entries = []gokeepasslib.Entry{entry}
	db.Content.Root.Groups = []gokeepasslib.Group{group}

	if err := db.LockProtectedEntries(); err != nil {
		t.Fatalf("LockProtectedEntries: %v", err)
	}
	f, err := os.Create(dbPath)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := gokeepasslib.NewEncoder(f).Encode(db); err != nil {
		f.Close()
		t.Fatalf("Encode: %v", err)
	}
	f.Close()

	m := NewManager()
	if !m.Open(dbPath, masterPassword) {
		t.Fatalf("Open failed: %s", m.LastError())
	}
	if got := m.Entries()[0].Password; got != "initial-pass!" {
		t.Fatalf("initial password = %q, want %q", got, "initial-pass!")
	}

	newPassword := "updated-p@ssw0rd-123!"
	if !m.UpdateEntry(0, "Email", "alice@example.com", newPassword, "https://mail.example.com", "updated note") {
		t.Fatalf("UpdateEntry failed: %s", m.LastError())
	}
	if err := m.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify in-memory state stays unlocked if rebuilt after Save.
	m.rebuildIndex()
	if got := m.Entries()[0].Password; got != newPassword {
		t.Errorf("password in memory after Save = %q, want %q", got, newPassword)
	}

	// Re-open the saved file and verify the updated password persisted.
	m2 := NewManager()
	if !m2.Open(dbPath, masterPassword) {
		t.Fatalf("second Open failed: %s", m2.LastError())
	}
	if len(m2.Entries()) != 1 {
		t.Fatalf("len(Entries) = %d, want 1", len(m2.Entries()))
	}
	if got := m2.Entries()[0].Password; got != newPassword {
		t.Errorf("password after reopen = %q, want %q", got, newPassword)
	}
}

