package zenuiraylib

import (
	"reflect"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"

	"github.com/ha1tch/zenui"
)

// frame is a keyboard for a test: the keys held, the keys that went down this frame, those repeating.
type frame struct {
	down, pressed, repeat []int32
	chars                 []rune
	clip                  string
	asked                 *int
}

func has(set []int32, k int32) bool {
	for _, v := range set {
		if v == k {
			return true
		}
	}
	return false
}

func (f frame) keyboard() keyboard {
	return keyboard{
		// a key that went down is held
		Down:    func(k int32) bool { return has(f.down, k) || has(f.pressed, k) },
		Pressed: func(k int32) bool { return has(f.pressed, k) },
		Repeat:  func(k int32) bool { return has(f.repeat, k) },
		Chars:   f.chars,
		Clipboard: func() string {
			if f.asked != nil {
				*f.asked++
			}
			return f.clip
		},
	}
}

func translate(f frame) (zenui.Input, string) {
	var in zenui.Input
	var copied string
	f.keyboard().translate(&in, func(s string) { copied = s })
	if in.SetClipboard != nil {
		in.SetClipboard("probe")
	}
	return in, copied
}

// Ctrl and Command are the same shortcut modifier, so the shortcuts of either system work on either.
func TestShortcutsAnswerToControlAndCommand(t *testing.T) {
	for _, mod := range []struct {
		name string
		key  int32
	}{{"Ctrl", rl.KeyLeftControl}, {"right Ctrl", rl.KeyRightControl}, {"Cmd", rl.KeyLeftSuper}, {"right Cmd", rl.KeyRightSuper}} {
		for _, c := range []struct {
			letter int32
			want   rune
		}{{rl.KeyC, 'c'}, {rl.KeyX, 'x'}, {rl.KeyA, 'a'}, {rl.KeyL, 'l'}} {
			in, _ := translate(frame{down: []int32{mod.key}, pressed: []int32{c.letter}})
			if !in.Ctrl || !reflect.DeepEqual(in.CtrlChars, []rune{c.want}) {
				t.Errorf("%s+%c: Ctrl %v CtrlChars %q", mod.name, c.want, in.Ctrl, string(in.CtrlChars))
			}
		}
		asked := 0
		in, _ := translate(frame{down: []int32{mod.key}, pressed: []int32{rl.KeyV}, clip: "from the clipboard", asked: &asked})
		if in.Paste != "from the clipboard" || asked != 1 {
			t.Errorf("%s+V: paste %q, clipboard read %d times", mod.name, in.Paste, asked)
		}
	}
	// A letter on its own is nothing to a widget that listens for shortcuts, and does not read the clipboard.
	asked := 0
	in, _ := translate(frame{pressed: []int32{rl.KeyV}, chars: []rune{'v'}, asked: &asked})
	if in.Ctrl || len(in.CtrlChars) != 0 || in.Paste != "" || asked != 0 {
		t.Errorf("a V typed: %+v, clipboard read %d times", in, asked)
	}
}

// The clipboard is only read for a paste, and a copy reaches the host's.
func TestClipboardOnlyForAPaste(t *testing.T) {
	asked := 0
	in, _ := translate(frame{down: []int32{rl.KeyLeftControl}, pressed: []int32{rl.KeyC}, asked: &asked})
	if asked != 0 || in.Paste != "" {
		t.Errorf("Ctrl+C read the clipboard %d times", asked)
	}
	var copied string
	var in2 zenui.Input
	frame{}.keyboard().translate(&in2, func(s string) { copied = s })
	in2.SetClipboard("hello")
	if copied != "hello" {
		t.Errorf("SetClipboard did not reach the host: %q", copied)
	}
}

// The PC's older pairs: Ctrl+Insert copies, Shift+Insert pastes, Shift+Delete cuts and does not delete.
func TestInsertAndDeletePairs(t *testing.T) {
	in, _ := translate(frame{down: []int32{rl.KeyLeftControl}, pressed: []int32{rl.KeyInsert}})
	if !reflect.DeepEqual(in.CtrlChars, []rune{'c'}) {
		t.Errorf("Ctrl+Insert: %q", string(in.CtrlChars))
	}
	asked := 0
	in, _ = translate(frame{down: []int32{rl.KeyLeftShift}, pressed: []int32{rl.KeyInsert}, clip: "pasted", asked: &asked})
	if in.Paste != "pasted" || in.Ctrl {
		t.Errorf("Shift+Insert: paste %q ctrl %v", in.Paste, in.Ctrl)
	}
	in, _ = translate(frame{down: []int32{rl.KeyLeftShift}, pressed: []int32{rl.KeyDelete}})
	if !reflect.DeepEqual(in.CtrlChars, []rune{'x'}) || len(in.Keys) != 0 {
		t.Errorf("Shift+Delete: CtrlChars %q keys %v, want a cut and no Delete", string(in.CtrlChars), in.Keys)
	}
	in, _ = translate(frame{pressed: []int32{rl.KeyDelete}})
	if !reflect.DeepEqual(in.Keys, []zenui.Key{zenui.KeyDelete}) {
		t.Errorf("Delete: %v", in.Keys)
	}
	// Insert alone does nothing
	in, _ = translate(frame{pressed: []int32{rl.KeyInsert}})
	if in.Paste != "" || len(in.CtrlChars) != 0 {
		t.Errorf("Insert alone: %+v", in)
	}
}

// AltGr is Ctrl and Alt on a PC; what it types is text, not a shortcut.
func TestAltGrTypesText(t *testing.T) {
	asked := 0
	in, _ := translate(frame{
		down: []int32{rl.KeyLeftControl, rl.KeyLeftAlt}, pressed: []int32{rl.KeyV},
		chars: []rune{'@'}, asked: &asked,
	})
	if in.Ctrl || in.Alt || !reflect.DeepEqual(in.Chars, []rune{'@'}) || len(in.CtrlChars) != 0 || in.Paste != "" || asked != 0 {
		t.Errorf("AltGr text: %+v, clipboard read %d times", in, asked)
	}
	// Alt on its own stays Alt, for going by word
	in, _ = translate(frame{down: []int32{rl.KeyLeftAlt}, pressed: []int32{rl.KeyLeft}})
	if !in.Alt || in.Ctrl || !reflect.DeepEqual(in.Keys, []zenui.Key{zenui.KeyLeft}) {
		t.Errorf("Alt+Left: %+v", in)
	}
}

// On a Mac Cmd+Left and Cmd+Right go to the ends of the line; Ctrl+Left stays a word jump.
func TestCommandArrowsGoToTheEnds(t *testing.T) {
	in, _ := translate(frame{down: []int32{rl.KeyLeftSuper}, pressed: []int32{rl.KeyLeft}})
	if !reflect.DeepEqual(in.Keys, []zenui.Key{zenui.KeyHome}) {
		t.Errorf("Cmd+Left: %v", in.Keys)
	}
	in, _ = translate(frame{down: []int32{rl.KeyRightSuper, rl.KeyLeftShift}, pressed: []int32{rl.KeyRight}})
	if !reflect.DeepEqual(in.Keys, []zenui.Key{zenui.KeyEnd}) || !in.Shift {
		t.Errorf("Cmd+Shift+Right: %v shift %v", in.Keys, in.Shift)
	}
	in, _ = translate(frame{down: []int32{rl.KeyLeftControl}, pressed: []int32{rl.KeyLeft}})
	if !reflect.DeepEqual(in.Keys, []zenui.Key{zenui.KeyLeft}) || !in.Ctrl {
		t.Errorf("Ctrl+Left: %v ctrl %v", in.Keys, in.Ctrl)
	}
}

// The editing and moving keys repeat while held; Enter, Escape and Tab act once.
func TestKeyRepeat(t *testing.T) {
	held := []int32{rl.KeyBackspace, rl.KeyDelete, rl.KeyLeft, rl.KeyRight, rl.KeyUp, rl.KeyDown, rl.KeyHome, rl.KeyEnd, rl.KeyPageUp, rl.KeyPageDown}
	in, _ := translate(frame{repeat: held})
	if len(in.Keys) != len(held) {
		t.Errorf("repeating editing keys gave %d keys, want %d: %v", len(in.Keys), len(held), in.Keys)
	}
	in, _ = translate(frame{repeat: []int32{rl.KeyEnter, rl.KeyKpEnter, rl.KeyEscape, rl.KeyTab}})
	if len(in.Keys) != 0 {
		t.Errorf("Enter, Escape and Tab repeated: %v", in.Keys)
	}
	in, _ = translate(frame{pressed: []int32{rl.KeyEnter, rl.KeyKpEnter, rl.KeyEscape, rl.KeyTab}})
	if len(in.Keys) != 4 {
		t.Errorf("Enter, Escape and Tab pressed: %v", in.Keys)
	}
	// Ctrl+V and Ctrl+X go on while held; Ctrl+A is a single act.
	asked := 0
	in, _ = translate(frame{down: []int32{rl.KeyLeftControl}, repeat: []int32{rl.KeyV, rl.KeyX, rl.KeyA}, clip: "x", asked: &asked})
	if in.Paste != "x" || !reflect.DeepEqual(in.CtrlChars, []rune{'v', 'x'}) {
		t.Errorf("held Ctrl+V/X/A: paste %q chars %q", in.Paste, string(in.CtrlChars))
	}
}
