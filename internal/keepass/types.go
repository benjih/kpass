package keepass

// Entry is a flat view of a KeePass entry for the QML list.
type Entry struct {
	Group    string
	Title    string
	Username string
	UUID     string
	Password string
	URL      string
	Notes    string
}
