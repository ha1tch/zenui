package zenui

import (
	"fmt"
	"net/url"
	"strings"
)

// What a dialog does with a typed or pasted path, and how it moves about: the
// listing of a folder, the recent files, resolving text to a place or a file,
// accepting, and completing with Tab.

// readDir lists dir, with the hidden entries when the filesystem can tell them.
func (d *Dialog) readDir(dir string) ([]Entry, error) {
	if h, ok := d.fs.(HiddenAware); ok {
		return h.ReadDirAll(dir)
	}
	return d.fs.ReadDir(dir)
}

// setDir loads dir's listing and resets selection and scroll.
func (d *Dialog) setDir(dir string) {
	es, err := d.readDir(dir)
	if err != nil {
		d.err = err.Error()
		return
	}
	d.err, d.hint, d.confirm = "", "", ""
	d.dir = dir
	d.container = "" // leaving any container
	d.recent = false
	d.raw = es
	d.rebuild()
	d.sel = -1
	d.scroll = 0
}

// setContainer enters a container file (e.g. a .zbun), listing its entries. The
// directory context (d.dir) is retained so goUp can climb back out.
func (d *Dialog) setContainer(p string) {
	es, err := d.fs.ReadContainer(p)
	if err != nil {
		d.err = err.Error()
		return
	}
	d.err, d.hint, d.confirm = "", "", ""
	d.container = p
	d.recent = false
	d.raw = es
	d.rebuild()
	d.sel = -1
	d.scroll = 0
}

// showRecent lists the recent files, filtered by the chosen set.
func (d *Dialog) showRecent() {
	d.err, d.hint, d.confirm = "", "", ""
	d.container = ""
	d.recent = true
	d.raw = d.recentEntries()
	d.rebuild()
	d.sel = -1
	d.scroll = 0
	if len(d.entries) == 0 {
		d.hint = "Nothing here yet"
	}
}

func (d *Dialog) recentEntries() []Entry {
	home := d.fs.UserHome()
	out := make([]Entry, 0, len(d.cfg.Recent))
	for _, p := range d.cfg.Recent {
		if p == "" {
			continue
		}
		folder := d.fs.Dir(p)
		if home != "" && (folder == home || strings.HasPrefix(folder, home+"/") || strings.HasPrefix(folder, home+"\\")) {
			folder = "~" + folder[len(home):]
		}
		out = append(out, Entry{Name: baseOf(p), Path: p, Detail: folder})
	}
	return out
}

// inContainer reports whether the dialog is currently browsing inside a
// container file rather than a directory.
func (d *Dialog) inContainer() bool { return d.container != "" }

// rebuild derives the list on show from the listing: what the view hides, what
// the file-type set leaves out, and the order.
func (d *Dialog) rebuild() {
	keep := d.selectedKey()
	var out []Entry
	switch {
	case d.inContainer():
		out = append(out, d.raw...)
		out = sortEntries(out)
	default:
		for _, e := range d.raw {
			if e.Hidden && !d.view.ShowHidden {
				continue
			}
			if d.recent {
				if !d.matches(e.Name) {
					continue
				}
			} else if !(e.IsDir || e.IsContainer || d.matches(e.Name)) {
				continue
			}
			out = append(out, e)
		}
		if !d.recent {
			sortView(out, d.view.Sort, d.view.Descending)
		}
	}
	d.entries = out
	d.infoChars = 0
	for _, e := range out {
		if n := len([]rune(e.Detail)); n > d.infoChars {
			d.infoChars = n
		}
	}
	d.sel = -1
	if keep != "" {
		for i, e := range d.entries {
			if entryKey(e) == keep {
				d.sel = i
				break
			}
		}
	}
	d.clampScroll()
	d.ensureVisible()
}

func entryKey(e Entry) string {
	if e.Path != "" {
		return e.Path
	}
	return e.Name
}

func (d *Dialog) selectedKey() string {
	if d.sel >= 0 && d.sel < len(d.entries) {
		return entryKey(d.entries[d.sel])
	}
	return ""
}

// canAccept reports whether the action button is currently enabled.
func (d *Dialog) canAccept() bool {
	text := strings.TrimSpace(d.name)
	if d.cfg.Mode == ModeSave {
		return text != ""
	}
	if d.typed && text != "" {
		return true
	}
	// Open: need a selected entry that is not a directory or a container (those
	// are for navigating into, not opening directly).
	if d.sel < 0 || d.sel >= len(d.entries) {
		return false
	}
	e := d.entries[d.sel]
	return !e.IsDir && !e.IsContainer
}

// setName puts text in the field with the caret at its end. typed says that it is
// the person's own text (an Open takes it as a path to resolve) rather than the
// echo of the selected entry.
func (d *Dialog) setName(s string, typed bool) {
	d.name = s
	d.typed = typed
	d.ed.toEnd(len([]rune(s)))
}

// selectIndex selects entry i, scrolls it into view and echoes its name in the
// field.
func (d *Dialog) selectIndex(i int) {
	if i < 0 || i >= len(d.entries) {
		return
	}
	d.sel = i
	d.ensureVisible()
	d.hint, d.confirm = "", ""
	e := d.entries[i]
	switch {
	case d.cfg.Mode == ModeSave:
		if !e.IsDir {
			d.setName(e.Name, true)
		}
	case e.IsDir || e.IsContainer:
		d.setName("", false)
	default:
		d.setName(e.Name, false)
	}
}

// --- accepting ---------------------------------------------------------------

// accept finalises the dialog with what the field and the selection say: the text
// the person typed, or the selected entry.
func (d *Dialog) accept() { d.commit() }

func (d *Dialog) commit() {
	if d.status != Active {
		return
	}
	d.err = ""
	text := strings.TrimSpace(d.name)
	if d.cfg.Mode == ModeOpen && !(d.typed && text != "") {
		d.acceptSelection()
		return
	}
	if text == "" {
		return
	}
	d.acceptText(text)
}

// finish ends the dialog with path, unless the host's Validate says no.
func (d *Dialog) finish(p string) bool {
	if v := d.cfg.Validate; v != nil {
		if msg := v(p); msg != "" {
			d.err = msg
			return false
		}
	}
	d.result = p
	d.status = Accepted
	return true
}

// pathOf is the path an entry of the list stands for. Inside a container it is
// "<containerpath>#<entryname>" so the caller can locate the entry within the
// archive.
func (d *Dialog) pathOf(e Entry) string {
	switch {
	case e.Path != "":
		return e.Path
	case d.inContainer():
		return d.container + "#" + e.Name
	}
	return d.fs.Join(d.dir, e.Name)
}

// acceptSelection accepts the selected file.
func (d *Dialog) acceptSelection() {
	if d.sel < 0 || d.sel >= len(d.entries) {
		return
	}
	e := d.entries[d.sel]
	if e.IsDir || e.IsContainer {
		return
	}
	p := d.pathOf(e)
	if e.Path != "" { // from the recent list: the file may be gone
		if t := d.lookup(p); t.kind != tFile {
			d.err = "No longer there: " + e.Name
			return
		}
	}
	d.finish(p)
}

// acceptText resolves what was typed and does what it names: goes into a folder,
// enters a bundle, opens a file, or saves.
func (d *Dialog) acceptText(text string) {
	if strings.ContainsAny(text, "*?") && !hasSep(text) && !d.inContainer() {
		d.setPattern(text)
		return
	}
	save := d.cfg.Mode == ModeSave
	if save && d.inContainer() {
		if hasSep(text) {
			d.err = "Inside a bundle, type an entry name"
			return
		}
		d.finish(d.container + "#" + strings.TrimSpace(text))
		return
	}
	t := d.resolveText(text)
	if save && (t.kind == tFile || t.kind == tMissing || t.kind == tContainer) {
		if p := d.withDefaultExt(t.path); p != t.path {
			t = d.lookup(p)
		}
	}
	if save && t.kind == tContainer {
		t.kind = tFile
	}
	switch t.kind {
	case tError:
		d.err = t.err
	case tDir:
		d.setDir(t.path)
		if d.err == "" {
			d.setName("", false)
		}
	case tContainer:
		d.setContainer(t.path)
		if d.err == "" {
			d.setName("", false)
		}
	case tEntry:
		d.finish(t.path)
	case tFile:
		if save && !d.cfg.NoOverwriteConfirm && d.confirm != t.path {
			d.confirm = t.path
			return
		}
		d.finish(t.path)
	case tMissing:
		if save {
			d.finish(t.path)
			return
		}
		// Not a name of anything, but the start of the name that type-ahead has
		// selected: the person means that one.
		if !hasSep(text) && d.sel >= 0 && d.sel < len(d.entries) && foldHasPrefix(d.entries[d.sel].Name, text) {
			d.openIndex(d.sel)
			return
		}
		if d.inContainer() {
			d.err = "No entry named " + t.path
		} else {
			d.err = "Not found: " + t.path
		}
	}
}

// openIndex does what Enter does on entry i of the list.
func (d *Dialog) openIndex(i int) {
	if i < 0 || i >= len(d.entries) {
		return
	}
	d.sel = i
	d.openSelection()
}

// openSelection enters a directory or container, or accepts a file.
func (d *Dialog) openSelection() {
	if d.sel < 0 || d.sel >= len(d.entries) {
		return
	}
	e := d.entries[d.sel]
	switch {
	case e.IsContainer && !d.inContainer():
		d.setContainer(d.fs.Join(d.dir, e.Name))
		if d.err == "" && d.cfg.Mode == ModeOpen {
			d.setName("", false)
		}
		return
	case e.IsDir:
		d.setDir(d.fs.Join(d.dir, e.Name))
		if d.err == "" && d.cfg.Mode == ModeOpen {
			d.setName("", false)
		}
		return
	}
	if d.cfg.Mode == ModeOpen {
		d.acceptSelection()
	} else {
		d.setName(e.Name, true)
	}
}

// goUp exits a container back to its directory, leaves the recent list, or
// navigates to the parent directory when already at directory level.
func (d *Dialog) goUp() {
	switch {
	case d.inContainer(), d.recent:
		d.setDir(d.dir)
	default:
		d.setDir(d.fs.Dir(d.dir))
	}
	if d.err == "" && d.cfg.Mode == ModeOpen {
		d.setName("", false)
	}
}

func hasExt(name string) bool {
	dot := strings.LastIndexByte(name, '.')
	slash := strings.LastIndexAny(name, "/\\")
	return dot > slash && dot < len(name)-1
}

// --- resolving text ------------------------------------------------------------

type targetKind int

const (
	tNone      targetKind = iota
	tDir                  // an existing folder
	tFile                 // an existing file
	tContainer            // an existing file that can be entered like a folder
	tEntry                // an entry of the container being browsed
	tMissing              // names nothing (yet), in a folder that exists
	tError                // cannot be resolved; err says why
)

// target is what a typed path comes to.
type target struct {
	kind  targetKind
	path  string // the full path (or "container#entry"); for a missing entry in a container, the name
	entry Entry
	err   string
}

func isAbsPath(s string) bool {
	if strings.HasPrefix(s, "/") || strings.HasPrefix(s, "\\") {
		return true
	}
	return len(s) >= 2 && s[1] == ':' && (s[0]|0x20 >= 'a' && s[0]|0x20 <= 'z')
}

func hasSep(s string) bool { return strings.ContainsAny(s, "/\\") }

// baseOf is the last element of a path, whichever separator it uses.
func baseOf(p string) string {
	p = strings.TrimRight(p, "/\\")
	if i := strings.LastIndexAny(p, "/\\"); i >= 0 {
		return p[i+1:]
	}
	return p
}

// normalizePasted undoes what copying a path usually adds: the spaces around it,
// the quotes Windows puts on "Copy as path", and a file:// address.
func normalizePasted(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && (s[0] == '"' || s[0] == '\'') && s[len(s)-1] == s[0] {
		s = strings.TrimSpace(s[1 : len(s)-1])
	}
	if strings.HasPrefix(s, "file://") {
		rest := s[len("file://"):]
		if i := strings.IndexByte(rest, '/'); i >= 0 { // drop a host name, keep the path
			rest = rest[i:]
		}
		if u, err := url.PathUnescape(rest); err == nil {
			rest = u
		}
		// file:///C:/dir is /C:/dir
		if len(rest) >= 3 && rest[0] == '/' && rest[2] == ':' {
			rest = rest[1:]
		}
		s = rest
	}
	return s
}

func (d *Dialog) expandHome(s string) string {
	if s == "~" {
		return d.fs.UserHome()
	}
	if strings.HasPrefix(s, "~/") || strings.HasPrefix(s, "~\\") {
		return d.fs.Join(d.fs.UserHome(), s[2:])
	}
	return s
}

// findEntry looks a name up in the list on show, exactly and then ignoring case.
func (d *Dialog) findEntry(name string) (Entry, bool) {
	for _, e := range d.entries {
		if e.Name == name {
			return e, true
		}
	}
	for _, e := range d.entries {
		if strings.EqualFold(e.Name, name) {
			return e, true
		}
	}
	return Entry{}, false
}

// resolveText says what a typed or pasted text stands for: a folder, a file, an
// entry of the bundle being browsed, or nothing. A text that is not absolute is
// taken from the folder on show; ~ is the home folder.
func (d *Dialog) resolveText(text string) target {
	s := normalizePasted(text)
	if s == "" {
		return target{kind: tNone}
	}
	if d.inContainer() {
		if hasSep(s) {
			return target{kind: tError, err: "Inside a bundle, type an entry name"}
		}
		if e, ok := d.findEntry(s); ok && !e.IsDir {
			return target{kind: tEntry, path: d.container + "#" + e.Name, entry: e}
		}
		return target{kind: tMissing, path: s}
	}
	trailing := strings.HasSuffix(s, "/") || strings.HasSuffix(s, "\\")
	s = d.expandHome(s)
	if !isAbsPath(s) {
		s = d.fs.Join(d.dir, s)
	}
	if a, err := d.fs.Abs(s); err == nil {
		s = a
	}
	t := d.lookup(s)
	if trailing {
		switch t.kind {
		case tFile, tContainer:
			return target{kind: tError, err: "Not a folder: " + baseOf(s)}
		case tMissing:
			return target{kind: tError, err: "No such folder: " + s}
		}
	}
	return t
}

// lookup finds p in the listing of its folder.
func (d *Dialog) lookup(p string) target {
	parent := d.fs.Dir(p)
	if parent == p { // the root of the filesystem
		return target{kind: tDir, path: p}
	}
	list, err := d.readDir(parent)
	if err != nil {
		return target{kind: tError, err: "No such folder: " + parent}
	}
	base := baseOf(p)
	found := -1
	for i, e := range list {
		if e.Name == base {
			found = i
			break
		}
	}
	if found < 0 { // a filesystem that ignores case finds it by another spelling
		for i, e := range list {
			if strings.EqualFold(e.Name, base) {
				if found >= 0 {
					found = -2
					break
				}
				found = i
			}
		}
	}
	if found < 0 {
		return target{kind: tMissing, path: p}
	}
	e := list[found]
	full := d.fs.Join(parent, e.Name)
	switch {
	case e.IsDir:
		return target{kind: tDir, path: full, entry: e}
	case e.IsContainer:
		return target{kind: tContainer, path: full, entry: e}
	}
	return target{kind: tFile, path: full, entry: e}
}

// typeAhead selects the first entry whose name starts with what is typed.
func (d *Dialog) typeAhead() {
	text := strings.TrimSpace(d.name)
	if d.cfg.Mode != ModeOpen || text == "" || hasSep(text) || strings.HasPrefix(text, "~") || strings.ContainsAny(text, "*?") {
		return
	}
	for i, e := range d.entries {
		if foldHasPrefix(e.Name, text) {
			d.sel = i
			d.ensureVisible()
			return
		}
	}
	d.sel = -1
}

// --- completion ---------------------------------------------------------------

// sepStyle is the separator to put after a folder name: the one the paths on show use.
func (d *Dialog) sepStyle() string {
	if strings.Contains(d.dir, "\\") && !strings.Contains(d.dir, "/") {
		return "\\"
	}
	return "/"
}

// complete finishes the name before the caret as far as it can be told: a single
// match is completed (a folder gets its separator), several are completed up to
// what they share and listed.
func (d *Dialog) complete() {
	runes := []rune(d.name)
	d.ed.clamp(len(runes))
	c := d.ed.caret
	head, tail := string(runes[:c]), string(runes[c:])
	dirPart, base := "", head
	if i := strings.LastIndexAny(head, "/\\"); i >= 0 {
		dirPart, base = head[:i+1], head[i+1:]
	}
	showDot := strings.HasPrefix(base, ".")

	var list []Entry
	switch {
	case d.inContainer():
		list = d.entries
	case dirPart == "" && !d.recent:
		list = d.entries
	default:
		dir := d.dir
		if dirPart != "" {
			dir = d.expandHome(dirPart)
			if !isAbsPath(dir) {
				dir = d.fs.Join(d.dir, dir)
			}
			if a, err := d.fs.Abs(dir); err == nil {
				dir = a
			}
		}
		es, err := d.readDir(dir)
		if err != nil {
			d.err = "No such folder: " + dir
			return
		}
		for _, e := range es {
			if e.Hidden && !d.view.ShowHidden && !showDot {
				continue
			}
			if e.IsDir || e.IsContainer || d.matches(e.Name) {
				list = append(list, e)
			}
		}
		sortView(list, d.view.Sort, d.view.Descending)
	}

	var cands []Entry
	for _, e := range list {
		if foldHasPrefix(e.Name, base) {
			cands = append(cands, e)
		}
	}
	d.err, d.hint, d.confirm = "", "", ""
	if len(cands) == 0 {
		d.hint = "No match"
		return
	}
	var comp string
	if len(cands) == 1 {
		comp = cands[0].Name
		if cands[0].IsDir && !d.inContainer() {
			comp += d.sepStyle()
		}
	} else {
		common := []rune(cands[0].Name)
		for _, e := range cands[1:] {
			er := []rune(e.Name)
			n := 0
			for n < len(common) && n < len(er) && strings.EqualFold(string(common[n]), string(er[n])) {
				n++
			}
			common = common[:n]
		}
		comp = string(common)
		if len([]rune(comp)) <= len([]rune(base)) {
			comp = base
			names := make([]string, 0, 6)
			for i, e := range cands {
				if i == 5 {
					names = append(names, "...")
					break
				}
				names = append(names, e.Name)
			}
			d.hint = fmt.Sprintf("%d matches: %s", len(cands), strings.Join(names, "  "))
		}
	}
	newHead := dirPart + comp
	d.name = newHead + tail
	d.typed = true
	n := len([]rune(newHead))
	d.ed.caret, d.ed.anchor, d.ed.set, d.ed.last = n, n, true, len([]rune(d.name))
	if d.cfg.Mode == ModeOpen {
		d.typeAhead()
	}
}
