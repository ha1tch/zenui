package zenui

import (
	"strings"
	"testing"
	"unicode/utf8"
)

const tfLH = 8

// The field is at (10,20), 120 wide and 12 high; its inner box is 2 in from each
// side, so the text starts at x=12 and the caret positions are 8 pixels apart.
var tfArea = Rect{X: 10, Y: 20, W: 120, H: 40}

// tfW measures a string as eight pixels a rune, like the test canvas.
func tfW(s string) int { return utf8.RuneCountInString(s) * 8 }

func tfWith(text string) *TextField {
	f := NewTextField()
	f.Text = text
	return f
}

func tfUpdate(f *TextField, in Input) TextEvent { return f.Update(in, tfArea, tfLH, tfW) }

func tfClick(f *TextField, x int) TextEvent {
	return tfUpdate(f, Input{MouseX: x, MouseY: tfArea.Y + 4, MousePressed: true})
}
func tfType(f *TextField, s string) TextEvent { return tfUpdate(f, Input{Chars: []rune(s)}) }
func tfKey(f *TextField, k ...Key) TextEvent  { return tfUpdate(f, Input{Keys: k}) }

// tfEnd starts an edit with the caret at the end of a short text.
func tfEnd(f *TextField) { tfClick(f, tfArea.X+tfArea.W-2) }

func TestATextFieldIgnoresInputUntilItIsClicked(t *testing.T) {
	f := tfWith("abc")
	if ev := tfType(f, "xyz"); ev != TextNone || f.Text != "abc" {
		t.Errorf("typing into a field never clicked gave %v and %q", ev, f.Text)
	}
	if ev := tfKey(f, KeyBackspace); ev != TextNone || f.Text != "abc" || f.Active() {
		t.Errorf("Backspace in a field never clicked gave %v and %q", ev, f.Text)
	}
	if ev := tfClick(f, 500); ev != TextNone || f.Active() {
		t.Error("a click outside the field started an edit")
	}
	if ev := tfClick(f, tfArea.X+5); ev != TextNone || !f.Active() {
		t.Errorf("a click inside gave %v, active %v, want an edit begun and no event", ev, f.Active())
	}
}

func TestTypingInsertsAtTheCaretAndReportsAChange(t *testing.T) {
	f := tfWith("ab")
	tfEnd(f)
	if ev := tfType(f, "cd"); ev != TextChanged || f.Text != "abcd" {
		t.Errorf("typing gave %v and %q, want a change and abcd", ev, f.Text)
	}
	tfKey(f, KeyLeft, KeyLeft)
	if ev := tfType(f, "X"); ev != TextChanged || f.Text != "abXcd" {
		t.Errorf("typing in the middle gave %v and %q, want abXcd", ev, f.Text)
	}
	tfType(f, "Y") // the caret moved on past what was typed
	if f.Text != "abXYcd" {
		t.Errorf("a second rune gave %q, want abXYcd", f.Text)
	}
	tfKey(f, KeyRight, KeyRight, KeyRight, KeyRight, KeyRight) // past the end: it stops there
	tfType(f, "!")
	if f.Text != "abXYcd!" {
		t.Errorf("after running off the end the text is %q, want abXYcd!", f.Text)
	}
}

func TestMaxLenAndTheFilterLimitWhatIsTyped(t *testing.T) {
	f := tfWith("ab")
	f.MaxLen = 4
	f.Accept = func(r rune) bool { return r >= 'a' && r <= 'z' }
	tfEnd(f)
	if ev := tfType(f, "1c2d3e"); ev != TextChanged || f.Text != "abcd" {
		t.Errorf("got %v and %q, want digits refused, the limit of 4 kept and abcd", ev, f.Text)
	}
	if ev := tfType(f, "z"); ev != TextNone || f.Text != "abcd" {
		t.Errorf("typing past the limit gave %v and %q, want no event and no change", ev, f.Text)
	}
}

func TestBackspaceDeletesBeforeTheCaret(t *testing.T) {
	f := tfWith("abc")
	tfEnd(f)
	if ev := tfKey(f, KeyBackspace); ev != TextChanged || f.Text != "ab" {
		t.Errorf("Backspace at the end gave %v and %q, want ab", ev, f.Text)
	}
	tfKey(f, KeyLeft)
	if tfKey(f, KeyBackspace); f.Text != "b" {
		t.Errorf("Backspace in the middle gave %q, want b", f.Text)
	}
	if ev := tfKey(f, KeyBackspace); ev != TextNone || f.Text != "b" {
		t.Errorf("Backspace at the start gave %v and %q, want nothing", ev, f.Text)
	}
}

func TestEnterCommitsAndEscapeRestores(t *testing.T) {
	f := tfWith("abc")
	tfEnd(f)
	tfType(f, "de")
	if ev := tfKey(f, KeyEnter); ev != TextCommitted || f.Text != "abcde" || f.Active() {
		t.Errorf("Enter gave %v, %q, active %v, want a commit of abcde and the edit over", ev, f.Text, f.Active())
	}
	tfEnd(f)
	tfType(f, "fg")
	if ev := tfKey(f, KeyEscape); ev != TextCancelled || f.Text != "abcde" || f.Active() {
		t.Errorf("Escape gave %v, %q, active %v, want the text from before the edit and the edit over", ev, f.Text, f.Active())
	}
}

func TestAClickAwayCommitsAndAClickInsidePlacesTheCaret(t *testing.T) {
	f := tfWith("abcd") // the caret positions are at x = 12, 20, 28, 36, 44
	tfClick(f, 29)      // nearest to 28: after "ab"
	tfType(f, "X")
	if f.Text != "abXcd" {
		t.Errorf("after a click at 29 typing gave %q, want abXcd", f.Text)
	}
	if ev := tfClick(f, 12+8*3+1); ev != TextNone || !f.Active() { // inside an active field: moves the caret
		t.Errorf("a click inside an active field gave %v, active %v", ev, f.Active())
	}
	tfType(f, "Z")
	if f.Text != "abXZcd" {
		t.Errorf("after moving the caret to 36 typing gave %q, want abXZcd", f.Text)
	}
	if ev := tfClick(f, 500); ev != TextCommitted || f.Active() || f.Text != "abXZcd" {
		t.Errorf("a click away gave %v, active %v, %q, want a commit", ev, f.Active(), f.Text)
	}
	if ev := tfClick(f, 500); ev != TextNone {
		t.Error("a second click away reported an edit that was not going on")
	}
}

func TestBlurCommitsOnlyAnEditThatIsGoingOn(t *testing.T) {
	f := tfWith("a")
	if ev := f.Blur(); ev != TextNone {
		t.Errorf("Blur of a field not being edited gave %v", ev)
	}
	tfEnd(f)
	tfType(f, "b")
	if ev := f.Blur(); ev != TextCommitted || f.Active() || f.Text != "ab" {
		t.Errorf("Blur gave %v, active %v, %q, want a commit of ab", ev, f.Active(), f.Text)
	}
}

func TestADisabledFieldCannotBeEditedAndDropsAnEditInProgress(t *testing.T) {
	f := tfWith("abc")
	f.Disabled = true
	if tfClick(f, tfArea.X+5); f.Active() {
		t.Error("a disabled field was activated")
	}
	f.Disabled = false
	tfEnd(f)
	f.Disabled = true
	if ev := tfType(f, "x"); ev != TextNone || f.Active() || f.Text != "abc" {
		t.Errorf("a field disabled mid-edit gave %v, active %v, %q", ev, f.Active(), f.Text)
	}
}

func TestRunesAreEditedWholeNotByByte(t *testing.T) {
	f := tfWith("añ")
	tfEnd(f)
	tfKey(f, KeyBackspace)
	if f.Text != "a" {
		t.Fatalf("Backspace over a two-byte rune left %q, want a", f.Text)
	}
	tfType(f, "ñb")
	tfKey(f, KeyLeft)
	tfKey(f, KeyBackspace)
	if f.Text != "ab" {
		t.Errorf("got %q, want ab", f.Text)
	}
}

func TestTheActiveFieldDrawsACaretAndALongTextScrollsToKeepItInView(t *testing.T) {
	th := DarkTheme()
	f := tfWith("abcd")
	draw := func() *listCanvas {
		cv := &listCanvas{}
		f.Draw(cv, th, tfArea, 1)
		return cv
	}
	caretAt := func(cv *listCanvas) (int, bool) {
		for _, fl := range cv.fills {
			if fl.C == th.Text {
				return fl.R.X, true
			}
		}
		return 0, false
	}
	cv := draw()
	if _, ok := caretAt(cv); ok {
		t.Error("an idle field drew a caret")
	}
	if len(cv.strokes) != 1 || cv.strokes[0].C != th.Border || cv.strokes[0].T != 1 {
		t.Errorf("an idle field's outline = %+v, want one thin border", cv.strokes)
	}
	tfEnd(f)
	cv = draw()
	if x, ok := caretAt(cv); !ok || x != 12+4*8 {
		t.Errorf("the caret is at %d (drawn %v), want 44, after the four runes", x, ok)
	}
	if len(cv.strokes) != 1 || cv.strokes[0].C != th.SelFill || cv.strokes[0].T != 2 {
		t.Errorf("an active field's outline = %+v, want a heavy one in the selection colour", cv.strokes)
	}

	f.Text = strings.Repeat("0123456789", 3) // 30 runes, 240 pixels, in a field with 116 inside
	f.Blur()
	tfClick(f, tfArea.X+5)
	for i := 0; i < 30; i++ {
		tfKey(f, KeyRight)
	}
	cv = draw()
	if len(cv.texts) != 1 || cv.texts[0].S != f.Text[16:] {
		t.Fatalf("the field shows %q, want the last 14 runes %q", cv.texts[0].S, f.Text[16:])
	}
	if x, ok := caretAt(cv); !ok || x != 12+14*8 || x >= tfArea.X+tfArea.W {
		t.Errorf("the caret is at %d, want 124, inside the field", x)
	}
}
