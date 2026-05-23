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
	mu      sync.Mutex
	manager = keepass.NewManager()
)

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

//export goCloseDatabase
func goCloseDatabase() {
	mu.Lock()
	defer mu.Unlock()
	manager.Close()
}

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

//export goGetGroupsJSON
func goGetGroupsJSON() *C.char {
	mu.Lock()
	defer mu.Unlock()
	b, _ := json.Marshal(manager.Groups())
	return C.CString(string(b))
}

//export goGetLastError
func goGetLastError() *C.char {
	mu.Lock()
	defer mu.Unlock()
	return C.CString(manager.LastError())
}

//export goDeleteEntry
func goDeleteEntry(index C.int) C.int {
	mu.Lock()
	defer mu.Unlock()
	if manager.DeleteEntry(int(index)) {
		return 1
	}
	return 0
}

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

//export goSaveDatabase
func goSaveDatabase() C.int {
	mu.Lock()
	defer mu.Unlock()
	if err := manager.Save(); err != nil {
		return 0
	}
	return 1
}

//export goFreeString
func goFreeString(s *C.char) {
	if s != nil {
		C.free(unsafe.Pointer(s))
	}
}

// NewDatabaseManager constructs the QML-facing DatabaseManager QObject.
func NewDatabaseManager() *qt6.QObject {
	ptr := C.newDatabaseManager(nil)
	return qt6.UnsafeNewQObject(unsafe.Pointer(ptr))
}
