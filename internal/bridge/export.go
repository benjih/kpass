// Package bridge is the cgo and C++ glue layer that connects QML to Go's
// KeePass logic. See internal/bridge/README.md for a full architecture walkthrough.
package bridge

/*
#cgo pkg-config: Qt6Core Qt6Gui
#include <stdlib.h>

void *newDatabaseManager(void *parent);
*/
import "C"

import (
	"encoding/json"
	"sync"
	"unsafe"

	"github.com/benjih/kpass/internal/keepass"
	qt6 "github.com/mappu/miqt/qt6"
)

var (
	// mu guards every manager call. Qt's event loop is single-threaded, but
	// the mutex makes the cgo boundary safe against any future concurrency.
	mu      sync.Mutex
	manager = keepass.NewManager()

	clipboardTimer    *qt6.QTimer
	clipboardLastText string
)

// goOpenDatabase opens the .kdbx file at path with the given password.
// Returns 1 on success, 0 on failure; retrieve the message with goGetLastError.
//
//export goOpenDatabase
func goOpenDatabase(path *C.char, password *C.char) C.int {
	mu.Lock()
	defer mu.Unlock()
	ok := manager.Open(C.GoString(path), C.GoString(password))
	if ok {
		return 1
	}
	return 0
}

// goCloseDatabase releases the currently open database from Go memory.
//
//export goCloseDatabase
func goCloseDatabase() {
	mu.Lock()
	defer mu.Unlock()
	manager.Close()
}

// goGetEntriesJSON returns all entries in the open database as a JSON array.
// Each element carries: group, title, username, uuid, password, url, notes.
// The caller owns the returned C string and must free it with goFreeString.
//
//export goGetEntriesJSON
func goGetEntriesJSON() *C.char {
	mu.Lock()
	defer mu.Unlock()
	type entryDTO struct {
		Group    string `json:"group"`
		Title    string `json:"title"`
		Username string `json:"username"`
		UUID     string `json:"uuid"`
		Password string `json:"password"`
		URL      string `json:"url"`
		Notes    string `json:"notes"`
	}
	dtos := make([]entryDTO, len(manager.Entries()))
	for i, e := range manager.Entries() {
		dtos[i] = entryDTO{
			Group:    e.Group,
			Title:    e.Title,
			Username: e.Username,
			UUID:     e.UUID,
			Password: e.Password,
			URL:      e.URL,
			Notes:    e.Notes,
		}
	}
	b, _ := json.Marshal(dtos)
	return C.CString(string(b))
}

// goGetGroupsJSON returns the deduplicated group names as a JSON string array.
// The caller owns the returned C string and must free it with goFreeString.
//
//export goGetGroupsJSON
func goGetGroupsJSON() *C.char {
	mu.Lock()
	defer mu.Unlock()
	b, _ := json.Marshal(manager.Groups())
	return C.CString(string(b))
}

// goGetLastError returns the last error message, or an empty string on success.
// The caller owns the returned C string and must free it with goFreeString.
//
//export goGetLastError
func goGetLastError() *C.char {
	mu.Lock()
	defer mu.Unlock()
	return C.CString(manager.LastError())
}

// goDeleteEntry removes the entry at index from the in-memory database.
// Returns 1 on success, 0 if the index is out of range.
//
//export goDeleteEntry
func goDeleteEntry(index C.int) C.int {
	mu.Lock()
	defer mu.Unlock()
	if manager.DeleteEntry(int(index)) {
		return 1
	}
	return 0
}

// goUpdateEntry overwrites the editable fields of the entry at index.
// Returns 1 on success, 0 if the index is out of range.
//
//export goUpdateEntry
func goUpdateEntry(index C.int, title, username, password, url, notes *C.char) C.int {
	mu.Lock()
	defer mu.Unlock()
	if manager.UpdateEntry(int(index), C.GoString(title), C.GoString(username),
		C.GoString(password), C.GoString(url), C.GoString(notes)) {
		return 1
	}
	return 0
}

// goSaveDatabase flushes the current in-memory state to the original .kdbx file.
// Returns 1 on success, 0 on write error; retrieve the message with goGetLastError.
//
//export goSaveDatabase
func goSaveDatabase() C.int {
	mu.Lock()
	defer mu.Unlock()
	if err := manager.Save(); err != nil {
		return 0
	}
	return 1
}

// goFreeString releases a C string that was allocated by C.CString in Go.
// C++ callers must invoke this for every string returned by a go* function.
//
//export goFreeString
func goFreeString(s *C.char) {
	if s != nil {
		C.free(unsafe.Pointer(s))
	}
}

// goCopyToClipboard sets the system clipboard to text and starts a 10-second
// timer that clears it — but only if the clipboard still holds that same text.
// Restarting before the timer fires resets the window.
//
//export goCopyToClipboard
func goCopyToClipboard(ctext *C.char) {
	text := C.GoString(ctext)

	mu.Lock()
	clipboardLastText = text
	mu.Unlock()

	qt6.QGuiApplication_Clipboard().SetText(text)

	if clipboardTimer == nil {
		clipboardTimer = qt6.NewQTimer()
		clipboardTimer.SetSingleShot(true)
		clipboardTimer.OnTimeout(func() {
			mu.Lock()
			last := clipboardLastText
			clipboardLastText = ""
			mu.Unlock()

			cb := qt6.QGuiApplication_Clipboard()
			if cb.Text() == last {
				cb.Clear()
			}
		})
	}
	clipboardTimer.Start(10_000)
}

// NewDatabaseManager allocates a DatabaseManager QObject and returns it as a
// *qt6.QObject so main.go can register it as a QML context property. The C++
// constructor is called here rather than in Go to let Qt's moc infrastructure
// manage the object's lifetime.
func NewDatabaseManager() *qt6.QObject {
	ptr := C.newDatabaseManager(nil)
	return qt6.UnsafeNewQObject(unsafe.Pointer(ptr))
}
