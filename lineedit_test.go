package zenui

import (
	"testing"
	"unicode/utf8"
)

// The editing every single-line box shares: caret, selection, word jumps, delete, copy, cut, paste.

// typed runs edit over text with the caret placed, and returns the text.
func editFrom(text string, caret, anchor int, in Input, lim editLimits) (string, *lineEdit, bool) {
	e := &lineEdit{caret: caret, anchor: anchor, set: true, last: len([]rune(text))}
	out, changed := e.edit(text, in, lim)
	return out, e, changed
}

func TestLineEditTypesAtTheCaret(t *testing.T) {
	got, e, changed := editFrom("abef", 2, 2, Input{Chars: []rune("cd")}, editLimits{})
	if got != "abcdef" || e.caret != 4 || e.anchor != 4 || !changed {
		t.Errorf("typed: %q caret %d anchor %d changed %v", got, e.caret, e.anchor, changed)
	}
	// An untouched box has its caret at the end.
	e2 := &lineEdit{}
	if got, _ := e2.edit("ab", Input{Chars: []rune("c")}, editLimits{}); got != "abc" {
		t.Errorf("a box with no place yet typed at %q", got)
	}
	// Control characters are not text.
	if got, _, ch := editFrom("ab", 2, 2, Input{Chars: []rune{'\n', 0x7f, 'x'}}, editLimits{}); got != "abx" || !ch {
		t.Errorf("control characters let in: %q", got)
	}
	// Ctrl held: a letter is a shortcut, not text.
	if got, _, ch := editFrom("ab", 2, 2, Input{Chars: []rune("c"), Ctrl: true}, editLimits{}); got != "ab" || ch {
		t.Errorf("a typed rune with Ctrl held went in: %q", got)
	}
}

func TestLineEditLimits(t *testing.T) {
	got, _, _ := editFrom("ab", 2, 2, Input{Chars: []rune("cdef")}, editLimits{MaxLen: 4})
	if got != "abcd" {
		t.Errorf("MaxLen 4: %q", got)
	}
	got, _, _ = editFrom("", 0, 0, Input{Chars: []rune("a1b2")}, editLimits{Accept: func(r rune) bool { return r >= '0' && r <= '9' }})
	if got != "12" {
		t.Errorf("Accept digits: %q", got)
	}
	// Typing over a selection frees its room.
	got, _, _ = editFrom("abcd", 4, 0, Input{Chars: []rune("xy")}, editLimits{MaxLen: 4})
	if got != "xy" {
		t.Errorf("typed over a full selection with MaxLen: %q", got)
	}
}

func TestLineEditBackspaceAndDeleteByRune(t *testing.T) {
	// A letter that is more than one byte goes whole.
	got, e, _ := editFrom("aé€b", 3, 3, Input{Keys: []Key{KeyBackspace}}, editLimits{})
	if got != "aéb" || e.caret != 2 || !utf8.ValidString(got) {
		t.Errorf("backspace over €: %q caret %d", got, e.caret)
	}
	got, e, _ = editFrom("aé€b", 1, 1, Input{Keys: []Key{KeyDelete}}, editLimits{})
	if got != "a€b" || e.caret != 1 {
		t.Errorf("delete of é: %q caret %d", got, e.caret)
	}
	// At the ends they do nothing and report no change.
	if got, _, ch := editFrom("ab", 0, 0, Input{Keys: []Key{KeyBackspace}}, editLimits{}); got != "ab" || ch {
		t.Errorf("backspace at the start: %q %v", got, ch)
	}
	if got, _, ch := editFrom("ab", 2, 2, Input{Keys: []Key{KeyDelete}}, editLimits{}); got != "ab" || ch {
		t.Errorf("delete at the end: %q %v", got, ch)
	}
}

func TestLineEditCaretMovesAndSelects(t *testing.T) {
	_, e, _ := editFrom("hello", 2, 2, Input{Keys: []Key{KeyRight}}, editLimits{})
	if e.caret != 3 || e.hasSel() {
		t.Errorf("right: caret %d sel %v", e.caret, e.hasSel())
	}
	_, e, _ = editFrom("hello", 2, 2, Input{Keys: []Key{KeyHome}}, editLimits{})
	if e.caret != 0 {
		t.Errorf("home: %d", e.caret)
	}
	_, e, _ = editFrom("hello", 2, 2, Input{Keys: []Key{KeyEnd}}, editLimits{})
	if e.caret != 5 {
		t.Errorf("end: %d", e.caret)
	}
	// Shift keeps the anchor.
	_, e, _ = editFrom("hello", 2, 2, Input{Keys: []Key{KeyRight, KeyRight}, Shift: true}, editLimits{})
	if lo, hi := e.sel(); lo != 2 || hi != 4 || e.caret != 4 {
		t.Errorf("shift+right twice: selection %d..%d caret %d", lo, hi, e.caret)
	}
	_, e, _ = editFrom("hello", 3, 3, Input{Keys: []Key{KeyHome}, Shift: true}, editLimits{})
	if lo, hi := e.sel(); lo != 0 || hi != 3 {
		t.Errorf("shift+home: %d..%d", lo, hi)
	}
	// An arrow with no Shift ends a selection at the edge it points to.
	_, e, _ = editFrom("hello", 4, 1, Input{Keys: []Key{KeyLeft}}, editLimits{})
	if e.caret != 1 || e.hasSel() {
		t.Errorf("left over a selection: caret %d sel %v", e.caret, e.hasSel())
	}
	_, e, _ = editFrom("hello", 1, 4, Input{Keys: []Key{KeyRight}}, editLimits{})
	if e.caret != 4 || e.hasSel() {
		t.Errorf("right over a selection: caret %d sel %v", e.caret, e.hasSel())
	}
	// The ends hold.
	_, e, _ = editFrom("ab", 0, 0, Input{Keys: []Key{KeyLeft}}, editLimits{})
	if e.caret != 0 {
		t.Errorf("left at the start: %d", e.caret)
	}
}

func TestLineEditWordJumps(t *testing.T) {
	const path = "/home/u/my tapes/game_1.tap"
	jump := func(from int, k Key, shift bool) int {
		_, e, _ := editFrom(path, from, from, Input{Keys: []Key{k}, Ctrl: true, Shift: shift}, editLimits{})
		return e.caret
	}
	n := len([]rune(path))
	if got := jump(n, KeyLeft, false); got != n-len("tap") {
		t.Errorf("ctrl+left from the end: %d, want %d", got, n-3)
	}
	if got := jump(n-len("tap"), KeyLeft, false); got != n-len("game_1.tap") {
		t.Errorf("ctrl+left again: %d (a dot ends a word, and game_1 is one word)", got)
	}
	if got := jump(0, KeyRight, false); got != len("/home") {
		t.Errorf("ctrl+right from the start: %d", got)
	}
	if got := jump(n, KeyRight, false); got != n {
		t.Errorf("ctrl+right at the end: %d", got)
	}
	_, e, _ := editFrom("one two", 7, 7, Input{Keys: []Key{KeyLeft}, Ctrl: true, Shift: true}, editLimits{})
	if lo, hi := e.sel(); lo != 4 || hi != 7 {
		t.Errorf("ctrl+shift+left selects %d..%d, want 4..7", lo, hi)
	}
	// Ctrl+Backspace and Ctrl+Delete take a word.
	got, _, _ := editFrom("one two", 7, 7, Input{Keys: []Key{KeyBackspace}, Ctrl: true}, editLimits{})
	if got != "one " {
		t.Errorf("ctrl+backspace: %q", got)
	}
	got, _, _ = editFrom("one two", 0, 0, Input{Keys: []Key{KeyDelete}, Ctrl: true}, editLimits{})
	if got != " two" {
		t.Errorf("ctrl+delete: %q", got)
	}
}

func TestLineEditSelectionIsReplacedAndDeleted(t *testing.T) {
	got, e, _ := editFrom("abcdef", 4, 1, Input{Chars: []rune("X")}, editLimits{})
	if got != "aXef" || e.caret != 2 || e.hasSel() {
		t.Errorf("typed over a selection: %q caret %d", got, e.caret)
	}
	got, e, _ = editFrom("abcdef", 1, 4, Input{Keys: []Key{KeyBackspace}}, editLimits{})
	if got != "aef" || e.caret != 1 {
		t.Errorf("backspace over a selection: %q caret %d", got, e.caret)
	}
	got, _, _ = editFrom("abcdef", 1, 4, Input{Keys: []Key{KeyDelete}}, editLimits{})
	if got != "aef" {
		t.Errorf("delete over a selection: %q", got)
	}
}

func TestLineEditCopyCutPasteSelectAll(t *testing.T) {
	var clip string
	set := func(s string) { clip = s }

	// Ctrl+A selects everything and Ctrl+C copies it, leaving the text alone.
	e := &lineEdit{caret: 2, anchor: 2, set: true}
	out, changed := e.edit("hello", Input{CtrlChars: []rune{'a', 'c'}, Ctrl: true, SetClipboard: set}, editLimits{})
	if out != "hello" || changed || clip != "hello" {
		t.Errorf("select all and copy: %q changed %v clipboard %q", out, changed, clip)
	}
	// Ctrl+X cuts a selection.
	e = &lineEdit{caret: 1, anchor: 3, set: true}
	out, changed = e.edit("abcd", Input{CtrlChars: []rune{'x'}, Ctrl: true, SetClipboard: set}, editLimits{})
	if out != "ad" || !changed || clip != "bc" || e.caret != 1 {
		t.Errorf("cut: %q changed %v clipboard %q caret %d", out, changed, clip, e.caret)
	}
	// With nothing selected a copy leaves the clipboard as it was.
	clip = "kept"
	e = &lineEdit{caret: 1, anchor: 1, set: true}
	e.edit("abcd", Input{CtrlChars: []rune{'c'}, Ctrl: true, SetClipboard: set}, editLimits{})
	if clip != "kept" {
		t.Errorf("a copy with no selection changed the clipboard: %q", clip)
	}
	// No clipboard hook: a copy is a no-op, not a crash.
	e = &lineEdit{caret: 0, anchor: 2, set: true}
	e.edit("abcd", Input{CtrlChars: []rune{'c', 'x'}, Ctrl: true}, editLimits{})

	// Paste goes in at the caret, over a selection, as one line.
	got, e2, _ := editFrom("abef", 2, 2, Input{Paste: "cd"}, editLimits{})
	if got != "abcdef" || e2.caret != 4 {
		t.Errorf("paste: %q caret %d", got, e2.caret)
	}
	got, _, _ = editFrom("abcdef", 1, 5, Input{Paste: "-"}, editLimits{})
	if got != "a-f" {
		t.Errorf("paste over a selection: %q", got)
	}
	got, _, _ = editFrom("", 0, 0, Input{Paste: "/home/u/tape.tap\r\n"}, editLimits{})
	if got != "/home/u/tape.tap" {
		t.Errorf("a pasted line end stayed: %q", got)
	}
	got, _, _ = editFrom("", 0, 0, Input{Paste: "a\tb\x00c\nd"}, editLimits{})
	if got != "a bcd" {
		t.Errorf("a tab and a control character in a paste: %q", got)
	}
	got, _, _ = editFrom("ab", 2, 2, Input{Paste: "cdefgh"}, editLimits{MaxLen: 5})
	if got != "abcde" {
		t.Errorf("paste with MaxLen: %q", got)
	}
	got, _, _ = editFrom("", 0, 0, Input{Paste: "a1b2"}, editLimits{Accept: func(r rune) bool { return r >= '0' && r <= '9' }})
	if got != "12" {
		t.Errorf("paste with Accept: %q", got)
	}
	if got, _, ch := editFrom("ab", 1, 1, Input{Paste: "\r\n"}, editLimits{}); got != "ab" || ch {
		t.Errorf("a paste of nothing but a line end changed the text: %q", got)
	}
}

// The box scrolls only as far as it must, and comes back when the text is short again.
func TestLineEditView(t *testing.T) {
	tw := func(s string) int { return utf8.RuneCountInString(s) * 10 }
	runes := []rune("0123456789abcdefghij") // 20 runes, 200 px; the box shows 8 runes
	e := &lineEdit{set: true}
	e.caret, e.anchor = 20, 20
	if s := e.view(runes, 80, tw); s != 12 {
		t.Errorf("caret at the end shows from %d, want 12", s)
	}
	e.caret, e.anchor = 10, 10
	if s := e.view(runes, 80, tw); s != 10 {
		t.Errorf("caret moved left to 10 shows from %d (it scrolls only until the caret is on show)", s)
	}
	e.caret, e.anchor = 0, 0
	if s := e.view(runes, 80, tw); s != 0 {
		t.Errorf("caret at the start shows from %d", s)
	}
	// The text got short: the box shows it from its start.
	e.scroll = 12
	e.caret, e.anchor = 3, 3
	if s := e.view([]rune("abc"), 80, tw); s != 0 {
		t.Errorf("a short text shows from %d, want 0", s)
	}
}

// A press places the caret, a drag selects, Shift extends, a release ends the drag.
func TestLineEditPointer(t *testing.T) {
	tw := func(s string) int { return utf8.RuneCountInString(s) * 10 }
	runes := []rune("hello world")
	box := Rect{X: 100, Y: 10, W: 300, H: 20}
	inner := Rect{X: 104, Y: 10, W: 292, H: 20}
	e := &lineEdit{}
	// A press at x = 104 + 34 is nearest the caret after 3 runes.
	if !e.pointer(runes, Input{MousePressed: true, MouseX: 138, MouseY: 15}, box, inner, tw) {
		t.Fatal("a press inside was not taken")
	}
	if e.caret != 3 || e.anchor != 3 || !e.drag {
		t.Errorf("press: caret %d anchor %d drag %v", e.caret, e.anchor, e.drag)
	}
	// Dragging to the right selects.
	e.pointer(runes, Input{MouseDown: true, MouseX: 184, MouseY: 15}, box, inner, tw)
	if lo, hi := e.sel(); lo != 3 || hi != 8 {
		t.Errorf("drag: selection %d..%d, want 3..8", lo, hi)
	}
	// Dragging out of the box to the right takes the selection to the end.
	e.pointer(runes, Input{MouseDown: true, MouseX: 900, MouseY: 400}, box, inner, tw)
	if e.caret != 11 {
		t.Errorf("drag far right: caret %d", e.caret)
	}
	// The release ends the drag: a later move with no button changes nothing.
	e.pointer(runes, Input{MouseX: 120, MouseY: 15}, box, inner, tw)
	e.pointer(runes, Input{MouseDown: true, MouseX: 120, MouseY: 15}, box, inner, tw)
	if e.caret != 11 || e.drag {
		t.Errorf("after the release: caret %d drag %v", e.caret, e.drag)
	}
	// Shift+press extends the selection from the anchor.
	e.caret, e.anchor = 2, 2
	e.pointer(runes, Input{MousePressed: true, MouseX: 174, MouseY: 15, Shift: true}, box, inner, tw)
	if lo, hi := e.sel(); lo != 2 || hi != 7 {
		t.Errorf("shift+press: selection %d..%d, want 2..7", lo, hi)
	}
	// A press outside is not taken and stops a drag.
	e.drag = true
	if e.pointer(runes, Input{MousePressed: true, MouseX: 5, MouseY: 5}, box, inner, tw) || e.drag {
		t.Error("a press outside was taken, or the drag went on")
	}
}

// The selection is drawn in the selection colours, and only when the box has the focus.
func TestLineEditDrawsSelectionAndCaret(t *testing.T) {
	th := DefaultTheme()
	r := &caretRenderer{}
	e := &lineEdit{caret: 4, anchor: 1, set: true}
	inner := Rect{X: 0, Y: 0, W: 400, H: 20}
	runes := []rune("abcdef")
	e.drawText(r, th, inner, runes, 0, 1, th.Text, true)
	var selFill bool
	for _, f := range r.fills {
		if f.col == th.SelFill && f.rect.W == 3*6 {
			selFill = true
		}
	}
	if !selFill {
		t.Errorf("no selection fill of three runes: %v", r.fills)
	}
	if _, ok := r.shown("bcd"); !ok {
		t.Errorf("the selected text was not drawn on its own: %v", r.texts)
	}
	if len(r.carets(th, r.LineHeight(1))) != 1 {
		t.Errorf("no caret: %v", r.fills)
	}
	r2 := &caretRenderer{}
	e.drawText(r2, th, inner, runes, 0, 1, th.Text, false)
	for _, f := range r2.fills {
		if f.col == th.SelFill {
			t.Error("a selection was drawn in a box with no focus")
		}
	}
	if _, ok := r2.shown("abcdef"); !ok || len(r2.carets(th, r2.LineHeight(1))) != 0 {
		t.Errorf("an unfocused box draws its text whole and no caret: %v", r2.texts)
	}
}
