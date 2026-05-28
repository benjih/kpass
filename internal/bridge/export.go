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

// goCreateDatabase creates a new .kdbx file at path with the given master
// password and opens it in the manager. Returns 1 on success, 0 on failure.
//
//export goCreateDatabase
func goCreateDatabase(path *C.char, password *C.char) C.int {
	mu.Lock()
	defer mu.Unlock()
	if manager.Create(C.GoString(path), C.GoString(password)) {
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
// Each element carries: group, title, username, uuid, password, url, notes, attachments.
// The caller owns the returned C string and must free it with goFreeString.
//
//export goGetEntriesJSON
func goGetEntriesJSON() *C.char {
	mu.Lock()
	defer mu.Unlock()
	type entryDTO struct {
		Group       string   `json:"group"`
		Title       string   `json:"title"`
		Username    string   `json:"username"`
		UUID        string   `json:"uuid"`
		Password    string   `json:"password"`
		URL         string   `json:"url"`
		Notes       string   `json:"notes"`
		Attachments []string `json:"attachments"`
	}
	dtos := make([]entryDTO, len(manager.Entries()))
	for i, e := range manager.Entries() {
		names := make([]string, len(e.Attachments))
		for j, a := range e.Attachments {
			names[j] = a.Name
		}
		dtos[i] = entryDTO{
			Group:       e.Group,
			Title:       e.Title,
			Username:    e.Username,
			UUID:        e.UUID,
			Password:    e.Password,
			URL:         e.URL,
			Notes:       e.Notes,
			Attachments: names,
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

// goGetAttachmentData returns the raw bytes of the named attachment for the
// entry at index. *outLen is set to the byte count. The caller owns the
// returned buffer and must free it with goFreeString. Returns nil on error.
//
//export goGetAttachmentData
func goGetAttachmentData(index C.int, filename *C.char, outLen *C.int) *C.char {
	mu.Lock()
	defer mu.Unlock()
	*outLen = 0
	data, err := manager.GetAttachmentData(int(index), C.GoString(filename))
	if err != nil {
		return nil
	}
	size := C.size_t(len(data))
	if size == 0 {
		size = 1 // non-nil signals success even for empty files
	}
	ptr := (*C.char)(C.malloc(size))
	if len(data) > 0 {
		copy(unsafe.Slice((*byte)(unsafe.Pointer(ptr)), len(data)), data)
	}
	*outLen = C.int(len(data))
	return ptr
}

// goAddAttachment reads dataLen bytes from data and attaches them under
// filename to the entry at index. Returns 1 on success, 0 on failure.
//
//export goAddAttachment
func goAddAttachment(index C.int, filename *C.char, data *C.char, dataLen C.int) C.int {
	mu.Lock()
	defer mu.Unlock()
	goData := C.GoBytes(unsafe.Pointer(data), dataLen)
	if manager.AddAttachment(int(index), C.GoString(filename), goData) {
		return 1
	}
	return 0
}

// goDeleteAttachment removes the named attachment from the entry at index.
// Returns 1 on success, 0 if the attachment was not found.
//
//export goDeleteAttachment
func goDeleteAttachment(index C.int, filename *C.char) C.int {
	mu.Lock()
	defer mu.Unlock()
	if manager.DeleteAttachment(int(index), C.GoString(filename)) {
		return 1
	}
	return 0
}

// NewDatabaseManager allocates a DatabaseManager QObject and returns it as a
// *qt6.QObject so main.go can register it as a QML context property. The C++
// constructor is called here rather than in Go to let Qt's moc infrastructure
// manage the object's lifetime.
func NewDatabaseManager() *qt6.QObject {
	ptr := C.newDatabaseManager(nil)
	return qt6.UnsafeNewQObject(unsafe.Pointer(ptr))
}
