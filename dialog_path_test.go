package zenui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// The Open and Save dialogs' name field: typed and pasted paths, completion,
// type-ahead, file types, hidden files, order, the recent files, and the other
// choices a host can make.

var dlgT0 = time.Date(2026, 1, 2, 3, 4, 0, 0, time.UTC)

func richFS() *memFS {
	return &memFS{
		home: "/home/u",
		dirs: map[string][]Entry{
			"/":     {{Name: "home", IsDir: true}},
			"/home": {{Name: "u", IsDir: true}},
			"/home/u": {
				{Name: "art", IsDir: true},
				{Name: "games", IsDir: true},
				{Name: ".config", IsDir: true, Hidden: true},
				{Name: ".secret.tap", Hidden: true, Size: 1, ModTime: dlgT0},
				{Name: "disk2.dsk", Size: 200000, ModTime: dlgT0.Add(1 * time.Hour)},
				{Name: "disk10.dsk", Size: 200000, ModTime: dlgT0},
				{Name: "Hero.TAP", Size: 6000, ModTime: dlgT0.Add(2 * time.Hour), Detail: "2 blocks"},
				{Name: "notes.txt", Size: 12, ModTime: dlgT0.Add(3 * time.Hour)},
				{Name: "hello.tap", Size: 3000, ModTime: dlgT0.Add(4 * time.Hour)},
				{Name: "game.zbun", IsContainer: true},
			},
			"/home/u/art":   {{Name: "enemy.zcut"}},
			"/home/u/games": {{Name: "zork.z80"}, {Name: "zebra.z80"}, {Name: "zoo", IsDir: true}},
		},
		containers: map[string][]Entry{
			"/home/u/game.zbun": {{Name: "knight"}, {Name: "goblin"}},
		},
	}
}

// plainFS is a filesystem that cannot list hidden entries.
type plainFS struct{ FS }

func dlgOpen(cfg DialogConfig) *Dialog {
	if cfg.FS == nil {
		cfg.FS = richFS()
	}
	if cfg.StartDir == "" {
		cfg.StartDir = "/home/u"
	}
	return NewDialog(cfg)
}

func dlgType(d *Dialog, s string) Status  { return d.Update(Input{Chars: []rune(s)}) }
func dlgKeys(d *Dialog, k ...Key) Status  { return d.Update(Input{Keys: k}) }
func dlgEnter(d *Dialog) Status           { return dlgKeys(d, KeyEnter) }
func dlgPaste(d *Dialog, s string) Status { return d.Update(Input{Paste: s}) }
func dlgCtrl(d *Dialog, c rune, k ...Key) Status {
	return d.Update(Input{Ctrl: true, CtrlChars: []rune{c}, Keys: k})
}

// dlgRepeat presses a key n times, a frame each.
func dlgRepeat(d *Dialog, k Key, n int) {
	for i := 0; i < n; i++ {
		dlgKeys(d, k)
	}
}
func dlgDraw(d *Dialog, w, h int) { d.Draw(noopRenderer{}, w, h, DefaultTheme()) }
func dlgClick(d *Dialog, x, y int, dt float32) Status {
	return d.Update(Input{MouseX: x, MouseY: y, MousePressed: true, DeltaTime: dt})
}
func dlgNames(d *Dialog) string {
	var n []string
	for _, e := range d.entries {
		n = append(n, e.Name)
	}
	return strings.Join(n, ",")
}

func TestOpenTypedPathsResolve(t *testing.T) {
	for _, c := range []struct{ typed, want string }{
		{"/home/u/art/enemy.zcut", "/home/u/art/enemy.zcut"}, // absolute
		{"art/enemy.zcut", "/home/u/art/enemy.zcut"},         // from the folder on show
		{"../u/art/enemy.zcut", "/home/u/art/enemy.zcut"},    // with ..
		{"~/art/enemy.zcut", "/home/u/art/enemy.zcut"},       // from home
		{"hero.tap", "/home/u/Hero.TAP"},                     // ignoring case
		{"  hello.tap  ", "/home/u/hello.tap"},               // spaces round it
	} {
		d := dlgOpen(DialogConfig{Mode: ModeOpen})
		dlgType(d, c.typed)
		if st := dlgEnter(d); st != Accepted || d.Result() != c.want {
			t.Errorf("%q: status %v result %q err %q, want %q", c.typed, st, d.Result(), d.err, c.want)
		}
	}
}

func TestOpenTypedFolderGoesThere(t *testing.T) {
	for _, typed := range []string{"/home/u/art", "art", "art/", "~/art/", "../u/art"} {
		d := dlgOpen(DialogConfig{Mode: ModeOpen})
		dlgType(d, typed)
		if st := dlgEnter(d); st != Active || d.Dir() != "/home/u/art" || d.name != "" || d.typed {
			t.Errorf("%q: status %v dir %q name %q typed %v err %q", typed, st, d.Dir(), d.name, d.typed, d.err)
		}
	}
	// The root of the filesystem is a folder too.
	d := dlgOpen(DialogConfig{Mode: ModeOpen})
	dlgType(d, "/")
	dlgEnter(d)
	if d.Dir() != "/" || d.err != "" {
		t.Errorf("/: dir %q err %q", d.Dir(), d.err)
	}
}

func TestOpenTypedPathErrors(t *testing.T) {
	for _, c := range []struct{ typed, errHas string }{
		{"nope.tap", "Not found"},
		{"/nowhere/x.tap", "No such folder"},
		{"hello.tap/", "Not a folder"},
		{"missing/", "No such folder"},
	} {
		d := dlgOpen(DialogConfig{Mode: ModeOpen})
		dlgType(d, c.typed)
		st := dlgEnter(d)
		if st != Active || !strings.Contains(d.err, c.errHas) || d.Dir() != "/home/u" {
			t.Errorf("%q: status %v dir %q err %q, want an error with %q", c.typed, st, d.Dir(), d.err, c.errHas)
		}
	}
	// Typing again drops the error.
	d := dlgOpen(DialogConfig{Mode: ModeOpen})
	dlgType(d, "nope")
	dlgEnter(d)
	dlgType(d, "x")
	if d.err != "" {
		t.Errorf("the error stayed after more typing: %q", d.err)
	}
}

func TestOpenPastedPaths(t *testing.T) {
	for _, c := range []struct{ pasted, want string }{
		{"/home/u/hello.tap", "/home/u/hello.tap"},
		{"\"/home/u/hello.tap\"", "/home/u/hello.tap"},    // Windows "Copy as path"
		{"'/home/u/hello.tap'", "/home/u/hello.tap"},      // quoted
		{"/home/u/hello.tap\n", "/home/u/hello.tap"},      // a trailing line end
		{"file:///home/u/hello.tap", "/home/u/hello.tap"}, // a file address
		{"file://localhost/home/u/hello.tap", "/home/u/hello.tap"},
	} {
		d := dlgOpen(DialogConfig{Mode: ModeOpen})
		dlgPaste(d, c.pasted)
		if st := dlgEnter(d); st != Accepted || d.Result() != c.want {
			t.Errorf("%q: status %v result %q err %q, want %q", c.pasted, st, d.Result(), d.err, c.want)
		}
	}
	if got := normalizePasted("file:///home/u/My%20Game.tap"); got != "/home/u/My Game.tap" {
		t.Errorf("a file address with a space: %q", got)
	}
	if got := normalizePasted("file:///C:/Games/a.tap"); got != "C:/Games/a.tap" {
		t.Errorf("a Windows file address: %q", got)
	}
}

func TestOpenInsideABundleTakesEntryNames(t *testing.T) {
	d := dlgOpen(DialogConfig{Mode: ModeOpen})
	dlgType(d, "game.zbun")
	dlgEnter(d)
	if !d.inContainer() {
		t.Fatalf("typing a bundle's name did not enter it: err %q", d.err)
	}
	dlgType(d, "KNIGHT")
	if st := dlgEnter(d); st != Accepted || d.Result() != "/home/u/game.zbun#knight" {
		t.Errorf("entry by name: %v %q %q", st, d.Result(), d.err)
	}
	d = dlgOpen(DialogConfig{Mode: ModeOpen, StartContainer: "/home/u/game.zbun"})
	dlgType(d, "art/x")
	dlgEnter(d)
	if d.Status() != Active || !strings.Contains(d.err, "entry name") {
		t.Errorf("a path inside a bundle: %v %q", d.Status(), d.err)
	}
	d = dlgOpen(DialogConfig{Mode: ModeOpen, StartContainer: "/home/u/game.zbun"})
	dlgType(d, "dragon")
	dlgEnter(d)
	if d.Status() != Active || !strings.Contains(d.err, "No entry named") {
		t.Errorf("a missing entry: %v %q", d.Status(), d.err)
	}
}

// The field echoes the selected file; editing it makes it the person's own text; a
// text that names nothing but starts the selected name takes that entry.
func TestOpenEchoTypeAheadAndPrefix(t *testing.T) {
	d := dlgOpen(DialogConfig{Mode: ModeOpen})
	if got := dlgNames(d); got != "art,games,disk2.dsk,disk10.dsk,game.zbun,hello.tap,Hero.TAP,notes.txt" {
		t.Fatalf("the listing (natural order, hidden left out): %s", got)
	}
	dlgRepeat(d, KeyDown, 3) // art, games, disk2.dsk
	if d.name != "disk2.dsk" || d.typed {
		t.Fatalf("the echo of the selection: %q typed %v", d.name, d.typed)
	}
	dlgKeys(d, KeyUp) // games, a folder: nothing to echo
	if d.name != "" {
		t.Errorf("a folder selected echoes %q", d.name)
	}
	dlgKeys(d, KeyDown)
	dlgKeys(d, KeyBackspace) // "disk2.ds": now the person's text
	if d.name != "disk2.ds" || !d.typed {
		t.Fatalf("after Backspace: %q typed %v", d.name, d.typed)
	}
	if st := dlgEnter(d); st != Accepted || d.Result() != "/home/u/disk2.dsk" {
		t.Errorf("a prefix of the selected name: %v %q %q", st, d.Result(), d.err)
	}

	d = dlgOpen(DialogConfig{Mode: ModeOpen})
	dlgType(d, "he")
	if d.sel != 5 { // hello.tap
		t.Errorf("type-ahead on \"he\" selected %d (%s)", d.sel, d.entries[max(d.sel, 0)].Name)
	}
	dlgType(d, "r")
	if d.sel != 6 || d.entries[d.sel].Name != "Hero.TAP" {
		t.Errorf("type-ahead on \"her\" selected %d", d.sel)
	}
	if st := dlgEnter(d); st != Accepted || d.Result() != "/home/u/Hero.TAP" {
		t.Errorf("Enter on \"her\": %v %q", st, d.Result())
	}
	d = dlgOpen(DialogConfig{Mode: ModeOpen})
	dlgType(d, "zzz")
	if d.sel != -1 {
		t.Errorf("type-ahead with no match left %d selected", d.sel)
	}
	// A typed name that is a prefix of nothing selected is an error, not an accident.
	if dlgEnter(d); d.Status() != Active || d.err == "" {
		t.Errorf("Enter on zzz: %v %q", d.Status(), d.err)
	}
}

func TestOpenAcceptButtonFollowsTheText(t *testing.T) {
	d := dlgOpen(DialogConfig{Mode: ModeOpen})
	if d.canAccept() {
		t.Error("an empty field with nothing selected can be accepted")
	}
	dlgType(d, "x")
	if !d.canAccept() {
		t.Error("typed text cannot be accepted")
	}
	dlgKeys(d, KeyBackspace)
	if d.canAccept() {
		t.Error("an emptied field can be accepted")
	}
}

func TestSaveTypedPaths(t *testing.T) {
	for _, c := range []struct{ typed, want string }{
		{"boss", "/home/u/boss.zcut"},               // the default extension
		{"boss.scr", "/home/u/boss.scr"},            // an extension of its own
		{"/home/u/art/new", "/home/u/art/new.zcut"}, // a path
		{"art/new", "/home/u/art/new.zcut"},         // relative
		{"~/art/new.scr", "/home/u/art/new.scr"},    // from home
		{"../u/boss", "/home/u/boss.zcut"},          // with ..
	} {
		d := dlgOpen(DialogConfig{Mode: ModeSave, DefaultExt: "zcut"})
		dlgType(d, c.typed)
		if st := dlgEnter(d); st != Accepted || d.Result() != c.want {
			t.Errorf("%q: %v %q err %q, want %q", c.typed, st, d.Result(), d.err, c.want)
		}
	}
	// A folder that is not there is said so.
	d := dlgOpen(DialogConfig{Mode: ModeSave})
	dlgType(d, "/nowhere/x")
	if dlgEnter(d); d.Status() != Active || !strings.Contains(d.err, "No such folder") {
		t.Errorf("a missing folder: %v %q", d.Status(), d.err)
	}
	// A name that is a folder goes into it.
	d = dlgOpen(DialogConfig{Mode: ModeSave, DefaultExt: "zcut"})
	dlgType(d, "art")
	if dlgEnter(d); d.Status() != Active || d.Dir() != "/home/u/art" || d.name != "" {
		t.Errorf("a folder's name: %v dir %q name %q", d.Status(), d.Dir(), d.name)
	}
}

func TestSaveAsksBeforeReplacing(t *testing.T) {
	d := dlgOpen(DialogConfig{Mode: ModeSave})
	dlgType(d, "hello.tap")
	if st := dlgEnter(d); st != Active || d.confirm != "/home/u/hello.tap" {
		t.Fatalf("an existing name: %v confirm %q", st, d.confirm)
	}
	if got := d.okLabel(); got != "REPLACE" {
		t.Errorf("the button reads %q while asking", got)
	}
	if line, _ := d.statusLine(DefaultTheme()); !strings.Contains(line, "hello.tap exists") {
		t.Errorf("the status line while asking: %q", line)
	}
	// Escape drops the question and keeps the dialog.
	if st := dlgKeys(d, KeyEscape); st != Active || d.confirm != "" {
		t.Errorf("Escape while asking: %v confirm %q", st, d.confirm)
	}
	if st := dlgKeys(d, KeyEscape); st != Cancelled {
		t.Errorf("a second Escape: %v", st)
	}
	// Enter twice replaces.
	d = dlgOpen(DialogConfig{Mode: ModeSave})
	dlgType(d, "hello.tap")
	dlgEnter(d)
	if st := dlgEnter(d); st != Accepted || d.Result() != "/home/u/hello.tap" {
		t.Errorf("confirming: %v %q", st, d.Result())
	}
	// Changing the name takes the question back.
	d = dlgOpen(DialogConfig{Mode: ModeSave})
	dlgType(d, "hello.tap")
	dlgEnter(d)
	dlgType(d, "2")
	if d.confirm != "" {
		t.Errorf("the question stayed after the name changed: %q", d.confirm)
	}
	// The button confirms as well.
	d = dlgOpen(DialogConfig{Mode: ModeSave})
	dlgDraw(d, 800, 600)
	dlgType(d, "hello.tap")
	dlgEnter(d)
	if st := dlgClick(d, d.okRect.X+2, d.okRect.Y+2, 0); st != Accepted {
		t.Errorf("clicking REPLACE: %v", st)
	}
	// The default extension counts: "hello" with .tap is the existing hello.tap.
	d = dlgOpen(DialogConfig{Mode: ModeSave, DefaultExt: "tap"})
	dlgType(d, "hello")
	if dlgEnter(d); d.confirm != "/home/u/hello.tap" {
		t.Errorf("hello + .tap: confirm %q", d.confirm)
	}
	// A host that does not want the question.
	d = dlgOpen(DialogConfig{Mode: ModeSave, NoOverwriteConfirm: true})
	dlgType(d, "hello.tap")
	if st := dlgEnter(d); st != Accepted {
		t.Errorf("NoOverwriteConfirm: %v", st)
	}
}

func TestValidateCanRefuse(t *testing.T) {
	var seen []string
	check := func(p string) string {
		seen = append(seen, p)
		if strings.HasSuffix(p, ".txt") {
			return "Not a Spectrum file"
		}
		return ""
	}
	d := dlgOpen(DialogConfig{Mode: ModeOpen, Validate: check})
	dlgType(d, "notes.txt")
	if st := dlgEnter(d); st != Active || d.err != "Not a Spectrum file" {
		t.Errorf("a refused file: %v %q", st, d.err)
	}
	dlgKeys(d, KeyBackspace, KeyBackspace, KeyBackspace, KeyBackspace)
	dlgType(d, "tap") // notes.tap does not exist: not found, not validated
	d = dlgOpen(DialogConfig{Mode: ModeOpen, Validate: check})
	dlgType(d, "hello.tap")
	if st := dlgEnter(d); st != Accepted {
		t.Errorf("an accepted file: %v %q", st, d.err)
	}
	// The selection is validated as well, and so is a Save.
	d = dlgOpen(DialogConfig{Mode: ModeOpen, Validate: check})
	for i, e := range d.entries {
		if e.Name == "notes.txt" {
			d.sel = i
		}
	}
	if d.accept(); d.Status() != Active || d.err == "" {
		t.Errorf("a refused selection: %v %q", d.Status(), d.err)
	}
	d = dlgOpen(DialogConfig{Mode: ModeSave, Validate: check})
	dlgType(d, "a.txt")
	if dlgEnter(d); d.Status() != Active || d.err != "Not a Spectrum file" {
		t.Errorf("a refused save: %v %q", d.Status(), d.err)
	}
	if len(seen) == 0 {
		t.Error("Validate was never asked")
	}
}

func TestTabCompletion(t *testing.T) {
	for _, c := range []struct {
		typed, want string
		hint        string // what the status hint must start with, if anything
	}{
		{"hel", "hello.tap", ""},
		{"ar", "art/", ""}, // a folder gets its separator
		{"di", "disk", ""}, // what two names share
		{"g", "game", ""},
		{"he", "he", "2 matches"},
		{"zzz", "zzz", "No match"},
		{"games/zor", "games/zork.z80", ""},
		{"games/zoo", "games/zoo/", ""},
		{"games/zo", "games/zo", "2 matches"},
		{"~/ar", "~/art/", ""},
		{"/home/u/gam", "/home/u/game", ""},
		{"HEL", "hello.tap", ""}, // case does not matter, the name's own spelling wins
	} {
		d := dlgOpen(DialogConfig{Mode: ModeOpen})
		dlgType(d, c.typed)
		dlgKeys(d, KeyTab)
		if d.name != c.want || !strings.HasPrefix(d.hint, c.hint) {
			t.Errorf("%q + Tab: %q hint %q, want %q hint %q", c.typed, d.name, d.hint, c.want, c.hint)
		}
	}
	// The caret goes to the end of what was completed, and the text after it is kept.
	d := dlgOpen(DialogConfig{Mode: ModeOpen})
	dlgType(d, "hel.old")
	dlgKeys(d, KeyHome)
	for i := 0; i < 3; i++ {
		dlgKeys(d, KeyRight)
	}
	dlgKeys(d, KeyTab)
	if d.name != "hello.tap.old" || d.ed.caret != len("hello.tap") {
		t.Errorf("completing before the caret: %q caret %d", d.name, d.ed.caret)
	}
	// A completed folder can be gone into with Enter.
	d = dlgOpen(DialogConfig{Mode: ModeOpen})
	dlgType(d, "ar")
	dlgKeys(d, KeyTab)
	if dlgEnter(d); d.Dir() != "/home/u/art" {
		t.Errorf("Tab then Enter: dir %q", d.Dir())
	}
	// The file type limits what completes; folders always complete.
	d = dlgOpen(DialogConfig{Mode: ModeOpen, Filters: []string{"tap"}})
	dlgType(d, "no")
	dlgKeys(d, KeyTab)
	if d.hint != "No match" {
		t.Errorf("a name the file type leaves out: %q", d.hint)
	}
	// Hidden files complete only when asked for with the dot or the switch.
	d = dlgOpen(DialogConfig{Mode: ModeOpen})
	dlgType(d, "/home/u/.sec")
	dlgKeys(d, KeyTab)
	if d.name != "/home/u/.secret.tap" {
		t.Errorf("a dot starts a hidden name: %q", d.name)
	}
	// Inside a bundle.
	d = dlgOpen(DialogConfig{Mode: ModeOpen, StartContainer: "/home/u/game.zbun"})
	dlgType(d, "kn")
	dlgKeys(d, KeyTab)
	if d.name != "knight" {
		t.Errorf("in a bundle: %q", d.name)
	}
}

func TestWildcardLimitsTheList(t *testing.T) {
	d := dlgOpen(DialogConfig{Mode: ModeOpen})
	dlgType(d, "*.tap")
	if st := dlgEnter(d); st != Active || d.name != "" || d.custom != "*.tap" {
		t.Fatalf("a pattern: %v name %q custom %q err %q", st, d.name, d.custom, d.err)
	}
	if got := dlgNames(d); got != "art,games,game.zbun,hello.tap,Hero.TAP" {
		t.Errorf("limited to the pattern (folders and bundles stay): %s", got)
	}
	if line, _ := d.statusLine(DefaultTheme()); line != "Showing *.tap" {
		t.Errorf("status line: %q", line)
	}
	dlgType(d, "*")
	dlgEnter(d)
	if d.custom != "" || len(d.entries) != 8 {
		t.Errorf("* takes the limit away: custom %q, %d entries", d.custom, len(d.entries))
	}
	// A Save does the same with a pattern, never saves one.
	d = dlgOpen(DialogConfig{Mode: ModeSave})
	dlgType(d, "*.dsk")
	if dlgEnter(d); d.Status() != Active || d.custom != "*.dsk" {
		t.Errorf("save + pattern: %v %q", d.Status(), d.custom)
	}
}

func TestFileTypeSets(t *testing.T) {
	sets := []FilterSet{
		{Name: "Tapes", Exts: []string{"tap"}, DefaultExt: "tap"},
		{Name: "Disks", Exts: []string{".DSK"}, DefaultExt: "dsk"},
	}
	d := dlgOpen(DialogConfig{Mode: ModeOpen, FilterSets: sets})
	if len(d.sets) != 3 || d.sets[2].Name != "All files" || !d.hasSelector() {
		t.Fatalf("sets: %+v", d.sets)
	}
	if got := dlgNames(d); got != "art,games,game.zbun,hello.tap,Hero.TAP" {
		t.Errorf("the first set: %s", got)
	}
	d.Update(Input{Ctrl: true, Keys: []Key{KeyTab}})
	if got := dlgNames(d); got != "art,games,disk2.dsk,disk10.dsk,game.zbun" || d.view.Filter != 1 {
		t.Errorf("Ctrl+Tab: %s (filter %d)", got, d.view.Filter)
	}
	d.Update(Input{Ctrl: true, Keys: []Key{KeyTab}})
	if len(d.entries) != 8 {
		t.Errorf("All files: %d entries", len(d.entries))
	}
	d.Update(Input{Ctrl: true, Shift: true, Keys: []Key{KeyTab}})
	if d.view.Filter != 1 {
		t.Errorf("Ctrl+Shift+Tab: filter %d", d.view.Filter)
	}
	d.Update(Input{Ctrl: true, Keys: []Key{KeyTab}})
	d.Update(Input{Ctrl: true, Keys: []Key{KeyTab}})
	if d.view.Filter != 0 {
		t.Errorf("Ctrl+Tab wraps round: filter %d", d.view.Filter)
	}
	// The set a Save is on decides the extension.
	s := dlgOpen(DialogConfig{Mode: ModeSave, FilterSets: sets, DefaultExt: "zcut", View: View{Filter: 1}})
	dlgType(s, "x")
	if dlgEnter(s); s.Result() != "/home/u/x.dsk" {
		t.Errorf("save on the Disks set: %q", s.Result())
	}
	s = dlgOpen(DialogConfig{Mode: ModeSave, FilterSets: sets, DefaultExt: "zcut", View: View{Filter: 2}})
	dlgType(s, "x")
	if dlgEnter(s); s.Result() != "/home/u/x.zcut" {
		t.Errorf("save on All files: %q", s.Result())
	}
	// No "All files" if the host does not want it; a view's filter is held in range.
	n := dlgOpen(DialogConfig{Mode: ModeOpen, FilterSets: sets, NoAllFiles: true, View: View{Filter: 9}})
	if len(n.sets) != 2 || n.view.Filter != 1 {
		t.Errorf("NoAllFiles: %d sets, filter %d", len(n.sets), n.view.Filter)
	}
	// The plain Filters list is one set and no selector.
	f := dlgOpen(DialogConfig{Mode: ModeOpen, Filters: []string{"tap"}})
	if f.hasSelector() || dlgNames(f) != "art,games,game.zbun,hello.tap,Hero.TAP" {
		t.Errorf("Filters: selector %v, %s", f.hasSelector(), dlgNames(f))
	}
	// A typed pattern is shown by the selector, and a chosen set takes it away.
	p := dlgOpen(DialogConfig{Mode: ModeOpen, FilterSets: sets})
	dlgType(p, "*.txt")
	dlgEnter(p)
	if p.filterLabel() != "*.txt" {
		t.Errorf("the selector shows %q", p.filterLabel())
	}
	p.setFilter(1)
	if p.custom != "" || p.filterLabel() != "Disks" {
		t.Errorf("choosing a set: custom %q label %q", p.custom, p.filterLabel())
	}
}

func TestFileTypePopup(t *testing.T) {
	sets := []FilterSet{{Name: "Tapes", Exts: []string{"tap"}}, {Name: "Disks", Exts: []string{"dsk"}}}
	d := dlgOpen(DialogConfig{Mode: ModeOpen, FilterSets: sets})
	dlgDraw(d, 800, 600)
	fr := d.filterRect
	if fr.W == 0 || d.nameRect.X+d.nameRect.W >= fr.X {
		t.Fatalf("the selector does not sit beside the field: field %+v selector %+v", d.nameRect, fr)
	}
	dlgClick(d, fr.X+4, fr.Y+4, 0)
	if !d.popup {
		t.Fatal("clicking the selector did not open the list")
	}
	dlgDraw(d, 800, 600)
	if len(d.popupRows) != 3 {
		t.Fatalf("the open list has %d rows, want 3", len(d.popupRows))
	}
	// Escape closes the list and nothing else.
	if st := dlgKeys(d, KeyEscape); st != Active || d.popup {
		t.Errorf("Escape on the list: %v popup %v", st, d.popup)
	}
	// Arrows and Enter choose.
	dlgClick(d, fr.X+4, fr.Y+4, 0)
	dlgKeys(d, KeyDown)
	dlgKeys(d, KeyEnter)
	if d.view.Filter != 1 || d.popup || d.Status() != Active {
		t.Errorf("choosing with the keys: filter %d popup %v", d.view.Filter, d.popup)
	}
	// A click on a row chooses it; a click elsewhere only closes the list.
	dlgClick(d, fr.X+4, fr.Y+4, 0)
	dlgDraw(d, 800, 600)
	row := d.popupRows[2]
	dlgClick(d, row.X+4, row.Y+2, 0)
	if d.view.Filter != 2 || d.popup {
		t.Errorf("clicking the third row: filter %d popup %v", d.view.Filter, d.popup)
	}
	dlgClick(d, fr.X+4, fr.Y+4, 0)
	if st := dlgClick(d, d.listRect.X+4, d.listRect.Y+4, 0); st != Active || d.popup || d.sel != -1 {
		t.Errorf("a click away from the list: %v popup %v sel %d", st, d.popup, d.sel)
	}
}

func TestHiddenFiles(t *testing.T) {
	d := dlgOpen(DialogConfig{Mode: ModeOpen})
	if strings.Contains(dlgNames(d), ".config") {
		t.Fatal("hidden files are listed from the start")
	}
	dlgCtrl(d, 'h')
	if names := dlgNames(d); !strings.Contains(names, ".config") || !strings.Contains(names, ".secret.tap") || !d.view.ShowHidden {
		t.Errorf("Ctrl+H: %s", names)
	}
	dlgCtrl(d, 'h')
	if strings.Contains(dlgNames(d), ".config") {
		t.Error("Ctrl+H again did not hide them")
	}
	// A typed hidden name is found either way.
	dlgType(d, ".secret.tap")
	if st := dlgEnter(d); st != Accepted || d.Result() != "/home/u/.secret.tap" {
		t.Errorf("a typed hidden name: %v %q %q", st, d.Result(), d.err)
	}
	// The selection survives the switch.
	d = dlgOpen(DialogConfig{Mode: ModeOpen})
	dlgRepeat(d, KeyDown, 3)
	want := d.entries[d.sel].Name
	dlgCtrl(d, 'h')
	if d.sel < 0 || d.entries[d.sel].Name != want {
		t.Errorf("the selection after the switch: %d", d.sel)
	}
	// The switch in the sidebar.
	d = dlgOpen(DialogConfig{Mode: ModeOpen})
	dlgDraw(d, 800, 600)
	dlgClick(d, d.hiddenRect.X+4, d.hiddenRect.Y+2, 0)
	if !d.view.ShowHidden {
		t.Error("clicking the switch did not show hidden files")
	}
	// A filesystem that cannot list them has no switch, and the view cannot ask for them.
	p := dlgOpen(DialogConfig{Mode: ModeOpen, FS: plainFS{richFS()}, View: View{ShowHidden: true}})
	dlgDraw(p, 800, 600)
	dlgCtrl(p, 'h')
	if p.view.ShowHidden || p.hiddenRect.W != 0 || strings.Contains(dlgNames(p), ".config") {
		t.Errorf("without HiddenAware: show %v rect %+v", p.view.ShowHidden, p.hiddenRect)
	}
}

func TestOrder(t *testing.T) {
	files := func(d *Dialog) string {
		var n []string
		for _, e := range d.entries {
			if !e.IsDir {
				n = append(n, e.Name)
			}
		}
		return strings.Join(n, ",")
	}
	for _, c := range []struct {
		name string
		view View
		want string
	}{
		{"name", View{}, "disk2.dsk,disk10.dsk,game.zbun,hello.tap,Hero.TAP,notes.txt"},
		{"name down", View{Descending: true}, "notes.txt,Hero.TAP,hello.tap,game.zbun,disk10.dsk,disk2.dsk"},
		{"size", View{Sort: SortSize}, "game.zbun,notes.txt,hello.tap,Hero.TAP,disk2.dsk,disk10.dsk"},
		{"size down", View{Sort: SortSize, Descending: true}, "disk10.dsk,disk2.dsk,Hero.TAP,hello.tap,notes.txt,game.zbun"},
		{"date", View{Sort: SortDate}, "game.zbun,disk10.dsk,disk2.dsk,Hero.TAP,notes.txt,hello.tap"},
		{"type", View{Sort: SortType}, "disk2.dsk,disk10.dsk,hello.tap,Hero.TAP,notes.txt,game.zbun"},
	} {
		d := dlgOpen(DialogConfig{Mode: ModeOpen, View: c.view})
		if got := files(d); got != c.want {
			t.Errorf("%s: %s, want %s", c.name, got, c.want)
		}
		// Folders stay first whichever way it goes.
		if !d.entries[0].IsDir || !d.entries[1].IsDir {
			t.Errorf("%s: folders are not first: %s", c.name, dlgNames(d))
		}
	}
	if !naturalLess("disk2", "disk10") || naturalLess("disk10", "disk2") || !naturalLess("a", "B") || !naturalLess("a1", "a01x") && false {
		t.Error("naturalLess")
	}
	if !naturalLess("a", "ab") || naturalLess("ab", "a") || naturalLess("x", "x") {
		t.Error("naturalLess on prefixes and equal names")
	}
	if humanSize(0) != "0 B" || humanSize(1023) != "1023 B" || humanSize(1536) != "1.5 KB" || humanSize(3<<20) != "3.0 MB" || humanSize(5<<30) != "5.0 GB" {
		t.Error("humanSize")
	}
}

func TestDetailsColumnsAndHeadClicks(t *testing.T) {
	d := dlgOpen(DialogConfig{Mode: ModeOpen, View: View{Details: true}})
	dlgDraw(d, 1000, 700)
	var kinds []colKind
	for _, c := range d.cols {
		kinds = append(kinds, c.kind)
	}
	if len(kinds) != 4 || kinds[0] != colName || kinds[1] != colInfo || kinds[2] != colSize || kinds[3] != colDate {
		t.Fatalf("columns on a wide dialog: %v", kinds)
	}
	if d.headRect.W == 0 || d.listRect.Y != d.headRect.Y+d.headRect.H {
		t.Errorf("the head does not sit over the list: %+v %+v", d.headRect, d.listRect)
	}
	// The columns tile the width.
	sum := 0
	for _, c := range d.cols {
		sum += c.w
	}
	if sum != d.headRect.W {
		t.Errorf("the columns are %d wide in a head %d wide", sum, d.headRect.W)
	}
	// Clicking a head sorts; clicking it again reverses.
	size := d.cols[2]
	dlgClick(d, size.rect.X+4, size.rect.Y+2, 0)
	if d.view.Sort != SortSize || d.view.Descending {
		t.Fatalf("a click on Size: %+v", d.view)
	}
	dlgClick(d, size.rect.X+4, size.rect.Y+2, 0)
	if d.view.Sort != SortSize || !d.view.Descending {
		t.Errorf("a second click: %+v", d.view)
	}
	date := d.cols[3]
	dlgClick(d, date.rect.X+4, date.rect.Y+2, 0)
	if d.view.Sort != SortDate || d.view.Descending {
		t.Errorf("a click on Modified: %+v", d.view)
	}
	// The Info column is not a sort key.
	info := d.cols[1]
	dlgClick(d, info.rect.X+4, info.rect.Y+2, 0)
	if d.view.Sort != SortDate {
		t.Errorf("the Info head sorted: %+v", d.view)
	}
	// A narrow dialog gives the name its room first.
	n := dlgOpen(DialogConfig{Mode: ModeOpen, View: View{Details: true}})
	dlgDraw(n, 500, 400)
	if len(n.cols) != 1 || n.cols[0].kind != colName {
		t.Errorf("columns on a narrow dialog: %+v", n.cols)
	}
	// Ctrl+D and the sidebar switch turn the details on and off.
	e := dlgOpen(DialogConfig{Mode: ModeOpen})
	dlgDraw(e, 1000, 700)
	if e.headRect.W != 0 {
		t.Error("details on from the start")
	}
	dlgCtrl(e, 'd')
	dlgDraw(e, 1000, 700)
	if !e.view.Details || e.headRect.W == 0 {
		t.Error("Ctrl+D did not turn the details on")
	}
	dlgClick(e, e.detailsRect.X+4, e.detailsRect.Y+2, 0)
	if e.view.Details {
		t.Error("the switch did not turn the details off")
	}
}

func TestRecentFiles(t *testing.T) {
	recent := []string{"/home/u/hello.tap", "/home/u/games/zork.z80", "/home/u/gone.tap", "/home/u/disk2.dsk"}
	d := dlgOpen(DialogConfig{Mode: ModeOpen, Recent: recent, FilterSets: []FilterSet{{Name: "Tapes", Exts: []string{"tap"}}}})
	if len(d.places) == 0 || d.places[0].Path != recentPlace || d.places[0].Label != "Recent" {
		t.Fatalf("the first place: %+v", d.places)
	}
	dlgDraw(d, 800, 600)
	dlgClick(d, d.placeRects[0].X+4, d.placeRects[0].Y+2, 0)
	if !d.recent || dlgNames(d) != "hello.tap,gone.tap" {
		t.Fatalf("the recent list (in order, limited by the file type): recent %v %s", d.recent, dlgNames(d))
	}
	if d.entries[0].Detail != "~" {
		t.Errorf("the folder shown: %q", d.entries[0].Detail)
	}
	// All files shows them all, with their folders.
	d.setFilter(1)
	if dlgNames(d) != "hello.tap,zork.z80,gone.tap,disk2.dsk" || d.entries[1].Detail != "~/games" {
		t.Errorf("all recent: %s (%q)", dlgNames(d), d.entries[1].Detail)
	}
	// A file that is gone is said so; one that is there opens, from wherever it is.
	dlgRepeat(d, KeyDown, 3) // gone.tap
	if dlgEnter(d); d.Status() != Active || !strings.Contains(d.err, "No longer there") {
		t.Errorf("a missing recent file: %v %q", d.Status(), d.err)
	}
	dlgRepeat(d, KeyUp, 1) // zork.z80
	if st := dlgEnter(d); st != Accepted || d.Result() != "/home/u/games/zork.z80" {
		t.Errorf("a recent file: %v %q %q", st, d.Result(), d.err)
	}
	// ".." leaves the list for the folder it was on.
	d = dlgOpen(DialogConfig{Mode: ModeOpen, Recent: recent})
	d.showRecent()
	d.goUp()
	if d.recent || d.Dir() != "/home/u" {
		t.Errorf("up from the recent list: recent %v dir %q", d.recent, d.Dir())
	}
	// Entering a folder leaves it too, and the list does not order by columns.
	d.showRecent()
	d.sortBy(SortSize)
	if dlgNames(d) != "hello.tap,zork.z80,gone.tap,disk2.dsk" {
		t.Errorf("the recent list was reordered: %s", dlgNames(d))
	}
	// An empty list says so.
	e := dlgOpen(DialogConfig{Mode: ModeOpen, Recent: []string{"/home/u/disk2.dsk"}, Filters: []string{"tap"}})
	e.showRecent()
	if len(e.entries) != 0 || e.hint == "" {
		t.Errorf("an empty recent list: %d entries, hint %q", len(e.entries), e.hint)
	}
	// Without recent files there is no entry for them.
	if n := dlgOpen(DialogConfig{Mode: ModeOpen}); len(n.places) > 0 && n.places[0].Path == recentPlace {
		t.Error("a Recent place with nothing recent")
	}
}

func TestPlacesAndLabels(t *testing.T) {
	d := dlgOpen(DialogConfig{Mode: ModeOpen, ExtraPlaces: []Place{{Label: "Games", Path: "/home/u/games"}}})
	last := d.places[len(d.places)-1]
	if last.Label != "Games" || len(d.places) != 3 {
		t.Errorf("extra places: %+v", d.places)
	}
	r := dlgOpen(DialogConfig{Mode: ModeOpen, Places: []Place{{Label: "Only", Path: "/home/u/art"}}})
	if len(r.places) != 1 || r.places[0].Label != "Only" {
		t.Errorf("replaced places: %+v", r.places)
	}
	none := dlgOpen(DialogConfig{Mode: ModeOpen, Places: []Place{}})
	if len(none.places) != 0 {
		t.Errorf("no places: %+v", none.places)
	}

	l := dlgOpen(DialogConfig{Mode: ModeOpen, Recent: []string{"/home/u/hello.tap"}, Labels: DialogLabels{
		Open: "Abrir", Cancel: "Cancelar", Recent: "Recientes", Hidden: "Ocultos", Details: "Detalles", OpenHint: "Escriba una ruta",
	}})
	rec := &textLog{}
	l.Draw(rec, 800, 600, DefaultTheme())
	for _, want := range []string{"ABRIR", "CANCELAR", "Recientes", "[ ] Ocultos", "[ ] Detalles", "Escriba una ruta"} {
		if !rec.has(want) {
			t.Errorf("the dialog does not draw %q: %v", want, rec.texts)
		}
	}
	s := dlgOpen(DialogConfig{Mode: ModeSave, Labels: DialogLabels{Save: "Guardar", SaveHint: "Nombre"}})
	rec = &textLog{}
	s.Draw(rec, 800, 600, DefaultTheme())
	if !rec.has("GUARDAR") || !rec.has("Nombre") {
		t.Errorf("save labels: %v", rec.texts)
	}
	// The defaults.
	rec = &textLog{}
	dlgOpen(DialogConfig{Mode: ModeOpen}).Draw(rec, 800, 600, DefaultTheme())
	for _, want := range []string{"OPEN", "CANCEL", "Type or paste a path", "8 items"} {
		if !rec.has(want) {
			t.Errorf("default text %q missing: %v", want, rec.texts)
		}
	}
	// The error shows in the status line.
	e := dlgOpen(DialogConfig{Mode: ModeOpen})
	dlgType(e, "nope")
	dlgEnter(e)
	rec = &textLog{}
	e.Draw(rec, 800, 600, DefaultTheme())
	if !rec.has("! Not found: /home/u/nope") {
		t.Errorf("the error is not in the status line: %v", rec.texts)
	}
}

// textLog is a renderer that remembers the text drawn.
type textLog struct {
	noopRenderer
	texts []string
}

func (t *textLog) DrawText(s string, x, y, scale int, c Colour) { t.texts = append(t.texts, s) }
func (t *textLog) has(s string) bool {
	for _, x := range t.texts {
		if x == s {
			return true
		}
	}
	return false
}

func TestViewIsKeptAndGivenBack(t *testing.T) {
	d := dlgOpen(DialogConfig{Mode: ModeOpen, FilterSets: []FilterSet{{Name: "T", Exts: []string{"tap"}}}})
	dlgCtrl(d, 'h')
	dlgCtrl(d, 'd')
	d.sortBy(SortSize)
	d.sortBy(SortSize)
	d.setFilter(1)
	v := d.View()
	want := View{ShowHidden: true, Details: true, Sort: SortSize, Descending: true, Filter: 1}
	if v != want {
		t.Fatalf("View() = %+v, want %+v", v, want)
	}
	again := dlgOpen(DialogConfig{Mode: ModeOpen, FilterSets: []FilterSet{{Name: "T", Exts: []string{"tap"}}}, View: v})
	if again.View() != want || !strings.Contains(dlgNames(again), ".config") {
		t.Errorf("a dialog started from the view: %+v %s", again.View(), dlgNames(again))
	}
}

func TestDoubleClickNeedsTheTime(t *testing.T) {
	click := func(cfg DialogConfig, gap float32) *Dialog {
		d := dlgOpen(cfg)
		dlgDraw(d, 800, 600)
		x, y := d.listRect.X+10, d.listRect.Y+d.rowH/2
		dlgClick(d, x, y, 0)
		dlgClick(d, x, y, gap)
		return d
	}
	if d := click(DialogConfig{Mode: ModeOpen}, 0.2); d.Dir() != "/home/u/art" {
		t.Errorf("two quick clicks: dir %q", d.Dir())
	}
	if d := click(DialogConfig{Mode: ModeOpen}, 0.8); d.Dir() != "/home/u" || d.sel != 0 {
		t.Errorf("two slow clicks: dir %q sel %d", d.Dir(), d.sel)
	}
	if d := click(DialogConfig{Mode: ModeOpen, DoubleClickSeconds: 1}, 0.8); d.Dir() != "/home/u/art" {
		t.Errorf("a longer limit: dir %q", d.Dir())
	}
	if d := click(DialogConfig{Mode: ModeOpen, DoubleClickSeconds: -1}, 100); d.Dir() != "/home/u/art" {
		t.Errorf("no limit: dir %q", d.Dir())
	}
	// A click on a row that is not the one just clicked does not open it.
	d := dlgOpen(DialogConfig{Mode: ModeOpen})
	dlgDraw(d, 800, 600)
	x := d.listRect.X + 10
	dlgClick(d, x, d.listRect.Y+d.rowH/2, 0)
	dlgClick(d, x, d.listRect.Y+d.rowH+d.rowH/2, 0.1)
	if d.Dir() != "/home/u" || d.sel != 1 {
		t.Errorf("two rows clicked in turn: dir %q sel %d", d.Dir(), d.sel)
	}
}

// The field is an edit box like the others: caret by mouse, selection, copy, paste.
func TestNameFieldEditsLikeATextField(t *testing.T) {
	d := dlgOpen(DialogConfig{Mode: ModeSave, Name: "abcdef"})
	if d.name != "abcdef" || d.ed.caret != 6 {
		t.Fatalf("the starting name: %q caret %d", d.name, d.ed.caret)
	}
	dlgDraw(d, 800, 600)
	in := d.nameInner()
	d.Update(Input{MousePressed: true, MouseDown: true, MouseX: in.X + 16*2 + 1, MouseY: in.Y + 4})
	if d.ed.caret != 2 {
		t.Fatalf("a click put the caret at %d, want 2", d.ed.caret)
	}
	dlgType(d, "X")
	if d.name != "abXcdef" {
		t.Errorf("typing at the caret: %q", d.name)
	}
	// Home, Shift+End, then typing replaces.
	dlgKeys(d, KeyHome)
	d.Update(Input{Keys: []Key{KeyEnd}, Shift: true})
	dlgType(d, "new")
	if d.name != "new" {
		t.Errorf("replacing a selection: %q", d.name)
	}
	// Select all, copy, paste elsewhere.
	var clip string
	d.Update(Input{Ctrl: true, CtrlChars: []rune{'a', 'c'}, SetClipboard: func(s string) { clip = s }})
	if clip != "new" {
		t.Errorf("copied %q", clip)
	}
	dlgKeys(d, KeyRight)
	dlgPaste(d, " file\r\n")
	if d.name != "new file" {
		t.Errorf("pasted: %q", d.name)
	}
	// Ctrl+L selects the whole field, so typing starts again.
	dlgCtrl(d, 'l')
	dlgType(d, "z")
	if d.name != "z" {
		t.Errorf("Ctrl+L then typing: %q", d.name)
	}
	// Delete and word movement.
	d = dlgOpen(DialogConfig{Mode: ModeSave, Name: "one two"})
	d.Update(Input{Keys: []Key{KeyBackspace}, Ctrl: true})
	if d.name != "one " {
		t.Errorf("Ctrl+Backspace: %q", d.name)
	}
	// A stray press away from the field does not move the caret.
	dlgClick(d, 0, 0, 0)
}

func TestKeysAroundTheList(t *testing.T) {
	d := dlgOpen(DialogConfig{Mode: ModeOpen, StartDir: "/home/u/art"})
	d.Update(Input{Alt: true, Keys: []Key{KeyUp}})
	if d.Dir() != "/home/u" {
		t.Errorf("Alt+Up: dir %q", d.Dir())
	}
	if d.sel != -1 {
		t.Errorf("Alt+Up moved the selection: %d", d.sel)
	}
	// Enter on a selected folder goes in, as it always did.
	dlgKeys(d, KeyDown)
	dlgEnter(d)
	if d.Dir() != "/home/u/art" || d.name != "" {
		t.Errorf("Enter on a folder: dir %q name %q", d.Dir(), d.name)
	}
	// In a Save a folder is selected, the name left, and Enter with no name goes in.
	s := dlgOpen(DialogConfig{Mode: ModeSave})
	dlgKeys(s, KeyDown)
	if dlgEnter(s); s.Dir() != "/home/u/art" {
		t.Errorf("Save: Enter on a folder: dir %q", s.Dir())
	}
	// Clicking a file in a Save puts its name in the field.
	dlgKeys(s, KeyDown)
	if s.name != "enemy.zcut" {
		t.Errorf("Save: selecting a file: %q", s.name)
	}
	// Navigating in an Open clears what was typed.
	o := dlgOpen(DialogConfig{Mode: ModeOpen})
	dlgType(o, "zzz")
	o.goUp()
	if o.name != "" || o.typed {
		t.Errorf("up with a typed name: %q typed %v", o.name, o.typed)
	}
}

func TestSizeAndScaleOptions(t *testing.T) {
	d := dlgOpen(DialogConfig{Mode: ModeOpen, Width: 500, Height: 350})
	dlgDraw(d, 1600, 1200)
	if d.bounds.W != 500 || d.bounds.H != 350 {
		t.Errorf("the panel is %dx%d, want 500x350", d.bounds.W, d.bounds.H)
	}
	big := dlgOpen(DialogConfig{Mode: ModeOpen})
	dlgDraw(big, 1600, 1200)
	if big.bounds.W != 760 || big.bounds.H != 560 {
		t.Errorf("the default panel is %dx%d", big.bounds.W, big.bounds.H)
	}
	s1 := dlgOpen(DialogConfig{Mode: ModeOpen, Scale: 1})
	dlgDraw(s1, 800, 600)
	s2 := dlgOpen(DialogConfig{Mode: ModeOpen})
	dlgDraw(s2, 800, 600)
	if s1.rowH >= s2.rowH || s1.rowH != 8+4 {
		t.Errorf("row heights at scale 1 and 2: %d, %d", s1.rowH, s2.rowH)
	}
}

// Every state draws without trouble, at several sizes.
func TestDialogDrawsInEveryState(t *testing.T) {
	sets := []FilterSet{{Name: "Tapes", Exts: []string{"tap"}}}
	cfgs := []DialogConfig{
		{Mode: ModeOpen},
		{Mode: ModeSave, Name: "abc"},
		{Mode: ModeOpen, FilterSets: sets, View: View{Details: true, ShowHidden: true}, Recent: []string{"/home/u/hello.tap"}},
		{Mode: ModeOpen, StartContainer: "/home/u/game.zbun", View: View{Details: true}},
	}
	for i, cfg := range cfgs {
		d := dlgOpen(cfg)
		for _, sz := range [][2]int{{300, 200}, {800, 600}, {1920, 1080}} {
			dlgDraw(d, sz[0], sz[1])
			d.popup = d.hasSelector()
			dlgDraw(d, sz[0], sz[1])
			d.popup = false
			d.confirm = "/home/u/x"
			dlgDraw(d, sz[0], sz[1])
			d.confirm = ""
			d.showRecent()
			dlgDraw(d, sz[0], sz[1])
			d.setDir("/home/u")
		}
		_ = i
	}
}

// The real filesystem: hidden entries marked, sizes and times, a link to a folder
// counted as a folder, and a typed path resolved against it.
func TestOSFSListsAndResolves(t *testing.T) {
	dir := t.TempDir()
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	must(os.WriteFile(filepath.Join(dir, "a.tap"), []byte("12345"), 0o644))
	must(os.WriteFile(filepath.Join(dir, ".hidden"), []byte("x"), 0o644))
	must(os.Mkdir(filepath.Join(dir, "sub"), 0o755))
	linked := os.Symlink(filepath.Join(dir, "sub"), filepath.Join(dir, "link")) == nil

	fs := OSFS{}
	all, err := fs.ReadDirAll(dir)
	must(err)
	byName := map[string]Entry{}
	for _, e := range all {
		byName[e.Name] = e
	}
	if !byName[".hidden"].Hidden || byName["a.tap"].Hidden {
		t.Errorf("hidden flags: %+v", byName)
	}
	if a := byName["a.tap"]; a.Size != 5 || a.ModTime.IsZero() || a.IsDir {
		t.Errorf("a.tap: %+v", a)
	}
	if !byName["sub"].IsDir {
		t.Error("sub is not a folder")
	}
	if linked && !byName["link"].IsDir {
		t.Error("a link to a folder is listed as a file")
	}
	plain, err := fs.ReadDir(dir)
	must(err)
	for _, e := range plain {
		if e.Hidden || e.Name == ".hidden" {
			t.Errorf("ReadDir lists %q", e.Name)
		}
	}

	d := NewDialog(DialogConfig{Mode: ModeOpen, StartDir: dir})
	dlgType(d, filepath.Join(dir, "a.tap"))
	if st := dlgEnter(d); st != Accepted || d.Result() != filepath.Join(dir, "a.tap") {
		t.Errorf("a typed path on the real filesystem: %v %q %q", st, d.Result(), d.err)
	}
	d = NewDialog(DialogConfig{Mode: ModeOpen, StartDir: dir})
	dlgType(d, "sub"+string(filepath.Separator))
	dlgEnter(d)
	if d.Dir() != filepath.Join(dir, "sub") {
		t.Errorf("a typed folder: %q (%s)", d.Dir(), d.err)
	}
	dlgType(d, "..")
	dlgEnter(d)
	if d.Dir() != dir {
		t.Errorf("..: %q, want %q", d.Dir(), dir)
	}
	// A hidden file typed by name is found; the real home is expanded.
	d = NewDialog(DialogConfig{Mode: ModeOpen, StartDir: dir})
	dlgType(d, ".hidden")
	if st := dlgEnter(d); st != Accepted {
		t.Errorf("a typed hidden file: %v %q", st, d.err)
	}
	d = NewDialog(DialogConfig{Mode: ModeSave, StartDir: dir, DefaultExt: "tap"})
	dlgType(d, "a")
	if dlgEnter(d); d.confirm != filepath.Join(dir, "a.tap") {
		t.Errorf("saving over a real file: confirm %q err %q", d.confirm, d.err)
	}
}
