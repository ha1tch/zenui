package zenui

import "testing"

// The editing a TextField has from the shared line editor: Delete, Home and End, selection by
// keyboard and by mouse, Ctrl+A, copy, cut and paste.

func tfIn(f *TextField, in Input) TextEvent { return tfUpdate(f, in) }

func TestTextFieldDeleteHomeEnd(t *testing.T) {
	f := tfWith("abcd")
	tfEnd(f)
	tfKey(f, KeyHome)
	if ev := tfKey(f, KeyDelete); ev != TextChanged || f.Text != "bcd" {
		t.Errorf("Home then Delete: %v %q", ev, f.Text)
	}
	tfKey(f, KeyEnd)
	tfType(f, "Z")
	if f.Text != "bcdZ" {
		t.Errorf("End then a rune: %q", f.Text)
	}
	if ev := tfKey(f, KeyEnd, KeyDelete); ev != TextNone || f.Text != "bcdZ" {
		t.Errorf("Delete at the end changed the text: %v %q", ev, f.Text)
	}
}

func TestTextFieldSelectionByKeysIsTypedOverAndDeleted(t *testing.T) {
	f := tfWith("abcdef")
	tfEnd(f)
	tfIn(f, Input{Keys: []Key{KeyLeft, KeyLeft}, Shift: true}) // "ef" selected
	if ev := tfType(f, "X"); ev != TextChanged || f.Text != "abcdX" {
		t.Errorf("typed over a selection: %v %q", ev, f.Text)
	}
	tfIn(f, Input{Keys: []Key{KeyHome}, Shift: true})
	if ev := tfKey(f, KeyBackspace); ev != TextChanged || f.Text != "" {
		t.Errorf("Backspace over a selection: %v %q", ev, f.Text)
	}
}

func TestTextFieldSelectAllCopyCutPaste(t *testing.T) {
	var clip string
	set := func(s string) { clip = s }
	f := tfWith("hello world")
	tfEnd(f)
	tfIn(f, Input{CtrlChars: []rune{'a', 'c'}, Ctrl: true, SetClipboard: set})
	if clip != "hello world" || f.Text != "hello world" {
		t.Errorf("select all and copy: clipboard %q text %q", clip, f.Text)
	}
	if ev := tfIn(f, Input{CtrlChars: []rune{'x'}, Ctrl: true, SetClipboard: set}); ev != TextChanged || f.Text != "" {
		t.Errorf("cut: %v %q", ev, f.Text)
	}
	if ev := tfIn(f, Input{Paste: "pasted\n"}); ev != TextChanged || f.Text != "pasted" {
		t.Errorf("paste: %v %q", ev, f.Text)
	}
	tfIn(f, Input{Keys: []Key{KeyHome}})
	tfIn(f, Input{Paste: ">> "})
	if f.Text != ">> pasted" {
		t.Errorf("paste at the start: %q", f.Text)
	}
	// A field that is not being edited does not take a paste.
	g := tfWith("keep")
	if ev := tfIn(g, Input{Paste: "no"}); ev != TextNone || g.Text != "keep" {
		t.Errorf("a field not being edited took a paste: %v %q", ev, g.Text)
	}
	// Paste obeys MaxLen.
	h := tfWith("ab")
	h.MaxLen = 4
	tfEnd(h)
	tfIn(h, Input{Paste: "cdefg"})
	if h.Text != "abcd" {
		t.Errorf("paste with MaxLen 4: %q", h.Text)
	}
}

// A press and a drag select with the mouse; Shift+press extends; Escape puts back the text from
// before the edit even after a paste.
func TestTextFieldMouseSelection(t *testing.T) {
	f := tfWith("abcdef") // carets at x = 12 + 8n
	tfClick(f, 12+8*1)    // before b
	tfIn(f, Input{MouseDown: true, MouseX: 12 + 8*4, MouseY: tfArea.Y + 4})
	if f.ed.caret != 4 || f.ed.anchor != 1 {
		t.Fatalf("drag: caret %d anchor %d, want 4 and 1", f.ed.caret, f.ed.anchor)
	}
	tfType(f, "-")
	if f.Text != "a-ef" {
		t.Errorf("typed over a dragged selection: %q", f.Text)
	}
	// The release ends the drag.
	tfIn(f, Input{MouseX: 12 + 8*0, MouseY: tfArea.Y + 4})
	tfIn(f, Input{MouseDown: true, MouseX: 12, MouseY: tfArea.Y + 4})
	if f.ed.caret != 2 {
		t.Errorf("a move after the release changed the caret to %d", f.ed.caret)
	}
	// Shift+click extends from the caret.
	g := tfWith("abcdef")
	tfClick(g, 12+8*2)
	tfIn(g, Input{MousePressed: true, MouseX: 12 + 8*5, MouseY: tfArea.Y + 4, Shift: true})
	if lo, hi := g.ed.sel(); lo != 2 || hi != 5 {
		t.Errorf("shift+click selected %d..%d, want 2..5", lo, hi)
	}
	// Escape restores the text before the edit.
	g.Text = "abcdef"
	tfIn(g, Input{Paste: "ZZ"})
	if ev := tfKey(g, KeyEscape); ev != TextCancelled || g.Text != "abcdef" {
		t.Errorf("Escape after a paste: %v %q", ev, g.Text)
	}
}

// The selection is drawn only while the field is being edited.
func TestTextFieldDrawsItsSelection(t *testing.T) {
	th := DefaultTheme()
	f := tfWith("abcdef")
	tfEnd(f)
	tfIn(f, Input{Keys: []Key{KeyLeft, KeyLeft}, Shift: true})
	r := &caretRenderer{}
	f.Draw(r, th, tfArea, 1)
	var sel bool
	for _, fl := range r.fills {
		if fl.col == th.SelFill && fl.rect.W == 2*6 {
			sel = true
		}
	}
	if !sel {
		t.Errorf("no selection fill of two runes: %v", r.fills)
	}
	f.Blur()
	r2 := &caretRenderer{}
	f.Draw(r2, th, tfArea, 1)
	for _, fl := range r2.fills {
		if fl.col == th.SelFill && fl.rect.W == 2*6 {
			t.Error("a field that is not being edited draws a selection")
		}
	}
}

// Alt, which is Option on a Mac, goes by word as Ctrl does.
func TestTextFieldAltGoesByWord(t *testing.T) {
	f := tfWith("one two three")
	tfEnd(f)
	tfIn(f, Input{Keys: []Key{KeyLeft}, Alt: true})
	if f.ed.caret != 8 {
		t.Errorf("Alt+Left from the end: caret %d, want 8", f.ed.caret)
	}
	tfIn(f, Input{Keys: []Key{KeyLeft}, Alt: true, Shift: true})
	if lo, hi := f.ed.sel(); lo != 4 || hi != 8 {
		t.Errorf("Alt+Shift+Left selected %d..%d, want 4..8", lo, hi)
	}
	tfIn(f, Input{Keys: []Key{KeyBackspace}, Alt: true})
	if f.Text != "one three" {
		t.Errorf("Backspace over the selection: %q", f.Text)
	}
	tfIn(f, Input{Keys: []Key{KeyRight}, Alt: true})
	if f.ed.caret != 9 {
		t.Errorf("Alt+Right: caret %d, want 9", f.ed.caret)
	}
}
