package zenuiraylib

import (
	rl "github.com/gen2brain/raylib-go/raylib"

	"github.com/ha1tch/zenui"
)

// keyEntry pairs a raylib key with the zenui key it stands for. repeat says that
// holding the key goes on giving it (the keys of editing and moving about), as
// opposed to the keys that act once however long they are held (Enter, Escape, Tab).
type keyEntry struct {
	rlKey  int32
	k      zenui.Key
	repeat bool
}

// keymap is fixed by the two APIs' own vocabularies, so it is data here rather
// than a chain of if-statements.
var keymap = []keyEntry{
	{rl.KeyEnter, zenui.KeyEnter, false},
	{rl.KeyKpEnter, zenui.KeyEnter, false},
	{rl.KeyEscape, zenui.KeyEscape, false},
	{rl.KeyBackspace, zenui.KeyBackspace, true},
	{rl.KeyDelete, zenui.KeyDelete, true},
	{rl.KeyUp, zenui.KeyUp, true},
	{rl.KeyDown, zenui.KeyDown, true},
	{rl.KeyLeft, zenui.KeyLeft, true},
	{rl.KeyRight, zenui.KeyRight, true},
	{rl.KeyPageUp, zenui.KeyPageUp, true},
	{rl.KeyPageDown, zenui.KeyPageDown, true},
	{rl.KeyTab, zenui.KeyTab, false},
	{rl.KeyHome, zenui.KeyHome, true},
	{rl.KeyEnd, zenui.KeyEnd, true},
}

// keyboard is one frame of the keyboard as the translation needs it, behind
// functions so that a test can stand in for raylib.
type keyboard struct {
	Down      func(key int32) bool // the key is held
	Pressed   func(key int32) bool // the key went down this frame
	Repeat    func(key int32) bool // the key is repeating, held past the first press
	Chars     []rune               // the printable runes typed this frame
	Clipboard func() string        // the clipboard's text; only called for a paste
}

// stroke is a key pressed this frame, with its repeats.
func (kb keyboard) stroke(key int32, repeat bool) bool {
	return kb.Pressed(key) || (repeat && kb.Repeat != nil && kb.Repeat(key))
}

// translate fills the keyboard half of a zenui.Input.
//
// The shortcut modifier, Input.Ctrl, is Control or Command (the Super key,
// which is Command on a Mac), so Ctrl+C, Ctrl+X, Ctrl+V and Ctrl+A are Cmd+C,
// Cmd+X, Cmd+V and Cmd+A on a Mac, and a person on either keeps the habits of
// the other. The older PC pairs work as well: Ctrl+Insert copies, Shift+Insert
// pastes and Shift+Delete cuts. On a Mac, Cmd+Left and Cmd+Right go to the
// start and the end of the line, as there; Option (Alt) goes by word.
//
// A rune that was typed this frame means the keys were typing, not a shortcut:
// AltGr, which a PC reports as Ctrl and Alt together, makes the @ and the € of
// many layouts, and that must reach the box as text.
func (kb keyboard) translate(in *zenui.Input, setClipboard func(string)) {
	ctrl := kb.Down(rl.KeyLeftControl) || kb.Down(rl.KeyRightControl)
	super := kb.Down(rl.KeyLeftSuper) || kb.Down(rl.KeyRightSuper)
	in.Shift = kb.Down(rl.KeyLeftShift) || kb.Down(rl.KeyRightShift)
	in.Alt = kb.Down(rl.KeyLeftAlt) || kb.Down(rl.KeyRightAlt)
	in.Chars = append(in.Chars, kb.Chars...)
	in.SetClipboard = setClipboard

	typing := len(kb.Chars) > 0
	if typing && (ctrl || super) {
		in.Alt = false // AltGr's own Alt
	}
	in.Ctrl = (ctrl || super) && !typing

	for _, m := range keymap {
		if !kb.stroke(m.rlKey, m.repeat) {
			continue
		}
		k := m.k
		switch {
		case super && !ctrl && k == zenui.KeyLeft:
			k = zenui.KeyHome
		case super && !ctrl && k == zenui.KeyRight:
			k = zenui.KeyEnd
		case k == zenui.KeyDelete && in.Shift && !in.Ctrl && !in.Alt:
			// Shift+Delete is cut, not a delete that selects
			in.CtrlChars = append(in.CtrlChars, 'x')
			continue
		}
		in.Keys = append(in.Keys, k)
	}

	if in.Ctrl {
		for c := rl.KeyA; c <= rl.KeyZ; c++ {
			if kb.stroke(int32(c), c != rl.KeyA) {
				in.CtrlChars = append(in.CtrlChars, rune('a'+c-rl.KeyA))
			}
		}
		if kb.stroke(rl.KeyInsert, false) {
			in.CtrlChars = append(in.CtrlChars, 'c') // Ctrl+Insert copies
		}
	}
	paste := in.Ctrl && kb.stroke(rl.KeyV, true)
	if !typing && in.Shift && !in.Ctrl && kb.stroke(rl.KeyInsert, true) {
		paste = true
	}
	if paste && kb.Clipboard != nil {
		in.Paste = kb.Clipboard()
	}
}
