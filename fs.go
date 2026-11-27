package zenui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Entry is one item in a directory listing.
type Entry struct {
	Name  string // base name, no path
	IsDir bool
	// IsContainer marks a file that can be descended into like a folder (e.g. an
	// archive). The dialog shows a hint and enters it on open, listing its
	// contents via FS.ReadContainer.
	IsContainer bool
	// Meta holds optional short "key: value" lines describing the entry, shown in
	// the preview pane (e.g. frame count, dimensions, a label). May be nil.
	Meta []MetaLine
	// Size is the file's size in bytes and ModTime when it was last changed; the
	// details view shows them (a zero ModTime means the filesystem did not say).
	Size    int64
	ModTime time.Time
	// Hidden marks an entry that is hidden by the system's convention (a name
	// that begins with a dot); the dialog lists it only when the person asks.
	// A filesystem that does not implement HiddenAware never lists one.
	Hidden bool
	// Detail is a short text of the host's own for the details view's Info
	// column (e.g. "48K, 2 blocks"). May be empty.
	Detail string
	// Path is the full path of the entry, for a listing that does not come from
	// one folder, such as the recent files. Empty in a folder's own listing.
	Path string
}

// MetaLine is one labelled fact about an entry, for the preview pane.
type MetaLine struct {
	Key   string
	Value string
}

// Place is a favourite/shortcut shown in the sidebar (Home, Desktop, etc.).
type Place struct {
	Label string // display name
	Path  string // absolute directory path
}

// FS is the filesystem the dialog browses. It is an interface so the dialog can
// be unit-tested against an in-memory tree and so hosts can sandbox it.
type FS interface {
	// ReadDir returns the entries of dir (absolute path). Implementations should
	// not include "." or ".."; the dialog synthesises the parent entry itself.
	ReadDir(dir string) ([]Entry, error)
	// Abs returns an absolute, cleaned form of path.
	Abs(path string) (string, error)
	// Join joins path elements with the OS separator.
	Join(elem ...string) string
	// Dir returns all but the last element of path (the parent directory).
	Dir(path string) string
	// UserHome returns a sensible starting directory.
	UserHome() string
	// Places returns the favourite/shortcut directories for the sidebar, in
	// display order. Entries whose paths do not exist may be included; the dialog
	// shows them but a click that fails to list simply reports the error.
	Places() []Place
	// IsContainer reports whether a file (given its full path) is a descendable
	// container the dialog can browse into, like an archive. Implementations that
	// do not support containers should always return false.
	IsContainer(path string) bool
	// ReadContainer lists the entries inside a container file. It is only called
	// for paths where IsContainer returned true.
	ReadContainer(path string) ([]Entry, error)
}

// ContainerHandler lets a host teach OSFS about descendable container files
// (archives such as a .zbun bundle) without zenui knowing their format. The
// host sets it on an OSFS instance before use.
type ContainerHandler interface {
	// IsContainer reports whether path names a container this handler understands.
	IsContainer(path string) bool
	// ReadContainer lists the entries inside the container at path.
	ReadContainer(path string) ([]Entry, error)
}

// HiddenAware is the optional interface of a filesystem that can list its hidden
// entries as well. The dialog uses it, when it is there, for the switch that shows
// hidden files: ReadDir keeps hiding them, ReadDirAll lists them with Entry.Hidden
// set, and the dialog applies the person's choice itself.
type HiddenAware interface {
	ReadDirAll(dir string) ([]Entry, error)
}

// OSFS is the real filesystem, backed by the os and path/filepath packages. Set
// Containers to enable browsing into archive files.
type OSFS struct {
	Containers ContainerHandler
}

// IsContainer delegates to the configured handler, if any.
func (fs OSFS) IsContainer(path string) bool {
	return fs.Containers != nil && fs.Containers.IsContainer(path)
}

// ReadContainer delegates to the configured handler.
func (fs OSFS) ReadContainer(path string) ([]Entry, error) {
	if fs.Containers == nil {
		return nil, fmt.Errorf("zenui: no container handler configured")
	}
	return fs.Containers.ReadContainer(path)
}

// ReadDir lists dir without its hidden entries (those whose names begin with a dot).
func (fs OSFS) ReadDir(dir string) ([]Entry, error) {
	all, err := fs.ReadDirAll(dir)
	if err != nil {
		return nil, err
	}
	out := all[:0]
	for _, e := range all {
		if !e.Hidden {
			out = append(out, e)
		}
	}
	return out, nil
}

// ReadDirAll lists dir with the hidden entries, marked, and each entry's size and
// modification time. A link to a folder counts as a folder.
func (fs OSFS) ReadDirAll(dir string) ([]Entry, error) {
	des, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	out := make([]Entry, 0, len(des))
	for _, de := range des {
		name := de.Name()
		e := Entry{Name: name, IsDir: de.IsDir(), Hidden: strings.HasPrefix(name, ".")}
		full := filepath.Join(dir, name)
		if de.Type()&os.ModeSymlink != 0 {
			if fi, err := os.Stat(full); err == nil && fi.IsDir() {
				e.IsDir = true
			}
		}
		if fi, err := de.Info(); err == nil && !e.IsDir {
			e.Size, e.ModTime = fi.Size(), fi.ModTime()
		}
		if !e.IsDir && fs.Containers != nil {
			e.IsContainer = fs.Containers.IsContainer(full)
		}
		out = append(out, e)
	}
	return out, nil
}

func (OSFS) Abs(path string) (string, error) { return filepath.Abs(path) }
func (OSFS) Join(elem ...string) string      { return filepath.Join(elem...) }
func (OSFS) Dir(path string) string          { return filepath.Dir(path) }

func (OSFS) UserHome() string {
	if h, err := os.UserHomeDir(); err == nil {
		return h
	}
	return "."
}

// Places returns the standard favourite directories that exist on this machine:
// Home and its common subfolders, plus the filesystem root. Non-existent
// candidates are skipped so the sidebar only lists usable destinations.
func (OSFS) Places() []Place {
	var out []Place
	home, _ := os.UserHomeDir()
	if home != "" {
		out = append(out, Place{Label: "Home", Path: home})
		for _, sub := range []string{"Desktop", "Documents", "Downloads"} {
			p := filepath.Join(home, sub)
			if fi, err := os.Stat(p); err == nil && fi.IsDir() {
				out = append(out, Place{Label: sub, Path: p})
			}
		}
	}
	root := "/"
	if vol := filepath.VolumeName(home); vol != "" {
		root = vol + string(filepath.Separator) // Windows drive root
	}
	out = append(out, Place{Label: "Root", Path: root})
	return out
}

// sortEntries orders a listing: directories first, then files, each
// alphabetically (case-insensitive). It mutates and returns the slice.
func sortEntries(es []Entry) []Entry {
	sort.SliceStable(es, func(i, j int) bool {
		if es[i].IsDir != es[j].IsDir {
			return es[i].IsDir // dirs before files
		}
		return strings.ToLower(es[i].Name) < strings.ToLower(es[j].Name)
	})
	return es
}

// matchFilter reports whether name passes the extension filter. An empty filter
// matches everything. Extensions are compared case-insensitively and may be
// given with or without a leading dot.
func matchFilter(name string, exts []string) bool {
	if len(exts) == 0 {
		return true
	}
	lname := strings.ToLower(name)
	for _, e := range exts {
		e = strings.ToLower(strings.TrimPrefix(e, "."))
		if strings.HasSuffix(lname, "."+e) {
			return true
		}
	}
	return false
}
