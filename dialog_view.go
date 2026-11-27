package zenui

import (
	"fmt"
	"path"
	"sort"
	"strings"
	"unicode/utf8"
)

// FilterSet is one choice in the dialog's file-type selector: a name to show and
// the extensions that pass. Directories and containers always pass, so the person
// can always get about.
type FilterSet struct {
	Name string   // shown in the selector, e.g. "Tapes and snapshots"
	Exts []string // extensions that pass, with or without the dot; empty passes everything
	// DefaultExt, if set, is the extension a Save adds to a typed name that has
	// none while this set is the one chosen. Empty falls back to
	// DialogConfig.DefaultExt.
	DefaultExt string
}

// SortKey is what the list is ordered by. Folders always come first.
type SortKey int

const (
	SortName SortKey = iota // by name, with the numbers in a name in numeric order (disk2 before disk10)
	SortSize
	SortDate
	SortType // by extension, then name
)

// View is the part of a dialog's state a host may want to remember between one
// use of the dialog and the next: the person's choices about how to look at a
// folder. Pass it in DialogConfig.View and read it back with Dialog.View.
type View struct {
	ShowHidden bool    // list the files and folders whose names begin with a dot
	Details    bool    // one row per entry with size and date columns
	Sort       SortKey // what the list is ordered by
	Descending bool    // reverse the order
	// Filter is the index of the chosen file-type set in the dialog's list of
	// sets: DialogConfig.FilterSets in order, then the "All files" set the dialog
	// adds at the end (unless NoAllFiles).
	Filter int
}

// DialogLabels are the words the dialog draws, for a host that wants its own
// (another language, a house style). An empty field keeps the default.
type DialogLabels struct {
	Open     string // the action button of an Open dialog (default "OPEN")
	Save     string // the action button of a Save dialog (default "SAVE")
	Cancel   string // default "CANCEL"
	Replace  string // the action button while a Save asks to replace a file (default "REPLACE")
	Recent   string // the sidebar entry for the recent files (default "Recent")
	AllFiles string // the selector's last set (default "All files")
	Hidden   string // the hidden-files switch (default "Hidden files")
	Details  string // the details switch (default "Details")
	Name     string // the Name column (default "Name")
	Size     string // the Size column (default "Size")
	Modified string // the date column (default "Modified")
	Folder   string // the folder column of the recent list (default "Folder")
	Info     string // the column of Entry.Detail (default "Info")
	OpenHint string // shown in the empty field of an Open (default "Type or paste a path")
	SaveHint string // shown in the empty field of a Save (default "File name")
}

func pick(v, def string) string {
	if v != "" {
		return v
	}
	return def
}

func (d *Dialog) okLabel() string {
	if d.confirm != "" {
		return pick(d.cfg.Labels.Replace, "REPLACE")
	}
	if d.cfg.Mode == ModeSave {
		return pick(d.cfg.Labels.Save, "SAVE")
	}
	return pick(d.cfg.Labels.Open, "OPEN")
}

// buildSets makes the list of file-type sets the selector offers.
func buildSets(cfg DialogConfig) []FilterSet {
	var sets []FilterSet
	switch {
	case len(cfg.FilterSets) > 0:
		sets = append(sets, cfg.FilterSets...)
		if !cfg.NoAllFiles {
			sets = append(sets, FilterSet{Name: pick(cfg.Labels.AllFiles, "All files")})
		}
	case len(cfg.Filters) > 0:
		// the single, unnamed set of the plain Filters list: no selector
		sets = []FilterSet{{Exts: cfg.Filters}}
	}
	return sets
}

// View returns the person's choices about how to look at the folder, for the
// host to keep and give back in DialogConfig.View next time.
func (d *Dialog) View() View { return d.view }

// activeExts are the extensions that pass now.
func (d *Dialog) activeExts() []string {
	if d.view.Filter >= 0 && d.view.Filter < len(d.sets) {
		return d.sets[d.view.Filter].Exts
	}
	return nil
}

// matches reports whether a file name passes the chosen set, or the pattern the
// person typed (such as *.tap).
func (d *Dialog) matches(name string) bool {
	if d.custom != "" {
		return globMatch(d.custom, name)
	}
	return matchFilter(name, d.activeExts())
}

func globMatch(pattern, name string) bool {
	ok, err := path.Match(strings.ToLower(pattern), strings.ToLower(name))
	return err == nil && ok
}

// hasSelector reports whether there is a choice of file types to offer.
func (d *Dialog) hasSelector() bool { return len(d.sets) >= 2 }

// filterLabel is what the selector shows: the chosen set, or the typed pattern.
func (d *Dialog) filterLabel() string {
	if d.custom != "" {
		return d.custom
	}
	if d.view.Filter >= 0 && d.view.Filter < len(d.sets) {
		return d.sets[d.view.Filter].Name
	}
	return ""
}

// setFilter chooses a file-type set and drops any typed pattern.
func (d *Dialog) setFilter(i int) {
	if len(d.sets) == 0 {
		return
	}
	d.view.Filter = clampInt(i, 0, len(d.sets)-1)
	d.custom = ""
	d.hint = ""
	d.rebuild()
}

// cycleFilter moves to the next (or previous) set, wrapping round.
func (d *Dialog) cycleFilter(delta int) {
	n := len(d.sets)
	if n < 2 {
		return
	}
	d.setFilter(((d.view.Filter+delta)%n + n) % n)
}

// setPattern limits the list to the names that match a typed pattern. The
// pattern * (or *.*) takes the limit away again.
func (d *Dialog) setPattern(p string) {
	d.custom = p
	if p == "*" || p == "*.*" { // matches everything: the same as no pattern
		d.custom = ""
	}
	d.hint = ""
	d.err = ""
	d.rebuild()
	d.setName("", false)
}

// defaultExt is the extension a Save adds to a name that has none.
func (d *Dialog) defaultExt() string {
	if d.view.Filter >= 0 && d.view.Filter < len(d.sets) && d.sets[d.view.Filter].DefaultExt != "" {
		return d.sets[d.view.Filter].DefaultExt
	}
	return d.cfg.DefaultExt
}

// withDefaultExt adds the default extension to the base name of p, if it has none.
func (d *Dialog) withDefaultExt(p string) string {
	ext := d.defaultExt()
	if ext == "" || hasExt(baseOf(p)) {
		return p
	}
	return p + "." + strings.TrimPrefix(ext, ".")
}

// hiddenSupported reports whether the filesystem can list hidden entries, so the
// switch is worth offering.
func (d *Dialog) hiddenSupported() bool {
	_, ok := d.fs.(HiddenAware)
	return ok
}

func (d *Dialog) toggleHidden() {
	if !d.hiddenSupported() {
		return
	}
	d.view.ShowHidden = !d.view.ShowHidden
	d.rebuild()
}

func (d *Dialog) toggleDetails() {
	d.view.Details = !d.view.Details
}

// sortBy orders the list by key; choosing the key already in force reverses it.
func (d *Dialog) sortBy(key SortKey) {
	if d.recent || d.inContainer() {
		return
	}
	if d.view.Sort == key {
		d.view.Descending = !d.view.Descending
	} else {
		d.view.Sort, d.view.Descending = key, false
	}
	d.rebuild()
}

// sortView orders a listing: folders first, then by key, ties by name.
func sortView(es []Entry, key SortKey, desc bool) {
	cmp := func(a, b Entry) int {
		switch key {
		case SortSize:
			if a.Size != b.Size {
				if a.Size < b.Size {
					return -1
				}
				return 1
			}
		case SortDate:
			if !a.ModTime.Equal(b.ModTime) {
				if a.ModTime.Before(b.ModTime) {
					return -1
				}
				return 1
			}
		case SortType:
			ea, eb := strings.ToLower(extOf(a.Name)), strings.ToLower(extOf(b.Name))
			if ea != eb {
				if ea < eb {
					return -1
				}
				return 1
			}
		}
		switch {
		case naturalLess(a.Name, b.Name):
			return -1
		case naturalLess(b.Name, a.Name):
			return 1
		}
		return 0
	}
	sort.SliceStable(es, func(i, j int) bool {
		a, b := es[i], es[j]
		if a.IsDir != b.IsDir {
			return a.IsDir
		}
		c := cmp(a, b)
		if desc {
			c = -c
		}
		return c < 0
	})
}

func extOf(name string) string {
	if i := strings.LastIndexByte(name, '.'); i > 0 {
		return name[i+1:]
	}
	return ""
}

// naturalLess orders names the way a person reads them: case does not matter and
// a run of digits counts by its value, so disk2 comes before disk10.
func naturalLess(a, b string) bool {
	a, b = strings.ToLower(a), strings.ToLower(b)
	digits := func(s string) int {
		i := 0
		for i < len(s) && s[i] >= '0' && s[i] <= '9' {
			i++
		}
		return i
	}
	for a != "" && b != "" {
		if i, j := digits(a), digits(b); i > 0 && j > 0 {
			na, nb := strings.TrimLeft(a[:i], "0"), strings.TrimLeft(b[:j], "0")
			if len(na) != len(nb) {
				return len(na) < len(nb)
			}
			if na != nb {
				return na < nb
			}
			a, b = a[i:], b[j:]
			continue
		}
		ra, sa := utf8.DecodeRuneInString(a)
		rb, sb := utf8.DecodeRuneInString(b)
		if ra != rb {
			return ra < rb
		}
		a, b = a[sa:], b[sb:]
	}
	return len(a) < len(b)
}

// humanSize writes a file size for a column: 512 B, 1.5 KB, 12.0 MB.
func humanSize(n int64) string {
	switch {
	case n < 1<<10:
		return fmt.Sprintf("%d B", n)
	case n < 1<<20:
		return fmt.Sprintf("%.1f KB", float64(n)/(1<<10))
	case n < 1<<30:
		return fmt.Sprintf("%.1f MB", float64(n)/(1<<20))
	}
	return fmt.Sprintf("%.1f GB", float64(n)/(1<<30))
}

const dateLayout = "2006-01-02 15:04"

// foldHasPrefix reports whether s starts with prefix, ignoring case.
func foldHasPrefix(s, prefix string) bool {
	return strings.HasPrefix(strings.ToLower(s), strings.ToLower(prefix))
}
