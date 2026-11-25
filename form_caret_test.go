package zenui

import (
	"testing"
	"unicode/utf8"
)

// caretRenderer remembers the fills and the texts a Form draws, with a text width of 6 pixels a
// rune at scale 1.
type caretRenderer struct {
	noopRenderer
	fills []caretFill
	texts []caretText
}

type caretFill struct {
	rect Rect
	col  Colour
}

type caretText struct {
	s string
	x int
}

func (c *caretRenderer) FillRect(r Rect, col Colour) { c.fills = append(c.fills, caretFill{r, col}) }
func (c *caretRenderer) DrawText(s string, x, y, scale int, col Colour) {
	c.texts = append(c.texts, caretText{s, x})
}
func (c *caretRenderer) TextWidth(s string, scale int) int {
	return utf8.RuneCountInString(s) * 6 * scale
}

// carets are the thin bars of the text colour, as high as a line, that a draw made.
func (c *caretRenderer) carets(th Theme, lh int) []Rect {
	var out []Rect
	for _, f := range c.fills {
		if f.col == th.Text && f.rect.W == max(1, lh/8) && f.rect.H == lh {
			out = append(out, f.rect)
		}
	}
	return out
}

func (c *caretRenderer) shown(s string) (int, bool) {
	for _, t := range c.texts {
		if t.s == s {
			return t.x, true
		}
	}
	return 0, false
}

// The text box that has the focus has a caret at the end of its text, where the next rune goes; it
// moves as text is typed; the box that has not the focus has none.
func TestFormTextFieldHasACaretAtTheEndOfItsText(t *testing.T) {
	th := DefaultTheme()
	name, other := "", "kept"
	f := NewForm(FormConfig{Fields: []FormField{
		{Label: "Name", Type: FieldText, Text: &name, MinWidthCells: 20},
		{Label: "Other", Type: FieldText, Text: &other, Row: 1, MinWidthCells: 20},
	}})
	last := -1
	for _, typed := range []string{"", "D", "DEL", "DELETE"} {
		name = typed
		r := &caretRenderer{}
		f.Draw(r, 800, 600, th)
		rec := f.FieldRect(0)
		lh := r.LineHeight(f.scale)
		cs := r.carets(th, lh)
		if len(cs) != 1 {
			t.Fatalf("with %q: %d carets, want 1 (the box without the focus has none)", typed, len(cs))
		}
		pad := f.gridSize / 4
		if want := rec.X + pad + r.TextWidth(typed, f.scale); cs[0].X != want {
			t.Errorf("with %q: the caret is at x=%d, want %d", typed, cs[0].X, want)
		}
		if cs[0].Y < rec.Y || cs[0].Y+cs[0].H > rec.Y+rec.H {
			t.Errorf("with %q: the caret %+v is not inside the box %+v", typed, cs[0], rec)
		}
		if typed != "" {
			if x, ok := r.shown(typed); !ok || x != rec.X+pad {
				t.Errorf("with %q: the text is drawn at x=%d (%v), the caret is worked out from %d", typed, x, ok, rec.X+pad)
			}
		}
		if cs[0].X <= last {
			t.Errorf("with %q: the caret did not move right", typed)
		}
		last = cs[0].X
	}
	// Tab gives the focus, and so the caret, to the other box
	f.Update(Input{Keys: []Key{KeyTab}})
	r := &caretRenderer{}
	f.Draw(r, 800, 600, th)
	cs := r.carets(th, r.LineHeight(f.scale))
	if len(cs) != 1 || cs[0].Y < f.FieldRect(1).Y || cs[0].Y > f.FieldRect(1).Y+f.FieldRect(1).H {
		t.Errorf("after Tab the caret is %v, want one in the second box %+v", cs, f.FieldRect(1))
	}
}

// A text that is longer than the box shows its end, with the caret in the box, where the typing is.
func TestFormTextFieldShowsTheEndOfALongText(t *testing.T) {
	th := DefaultTheme()
	name := ""
	for i := 0; i < 60; i++ {
		name += string(rune('a' + i%26))
	}
	f := NewForm(FormConfig{Fields: []FormField{{Label: "Name", Type: FieldText, Text: &name, MinWidthCells: 12}}})
	r := &caretRenderer{}
	f.Draw(r, 800, 600, th)
	rec := f.FieldRect(0)
	cs := r.carets(th, r.LineHeight(f.scale))
	if len(cs) != 1 || cs[0].X+cs[0].W > rec.X+rec.W || cs[0].X < rec.X {
		t.Fatalf("a long text: carets %v, box %+v", cs, rec)
	}
	var shown string
	for _, tx := range r.texts {
		if len(tx.s) > 0 && len(tx.s) < len(name) && name[len(name)-len(tx.s):] == tx.s {
			shown = tx.s
		}
	}
	if shown == "" {
		t.Fatalf("the box shows no end of the text: %v", r.texts)
	}
	if r.TextWidth(shown, f.scale) > rec.W {
		t.Errorf("the text shown is %d wide in a box %d wide", r.TextWidth(shown, f.scale), rec.W)
	}
}

// Backspace takes away a whole rune, not a byte of it.
func TestFormTextFieldBackspaceRemovesAWholeRune(t *testing.T) {
	for _, c := range []struct{ in, want string }{
		{"año", "añ"}, {"añ", "a"}, {"日本", "日"}, {"é", ""}, {"a😀", "a"}, {"a", ""},
	} {
		name := c.in
		f := NewForm(FormConfig{Fields: []FormField{{Label: "Name", Type: FieldText, Text: &name}}})
		f.Draw(noopRenderer{}, 800, 600, DefaultTheme())
		f.Update(Input{Keys: []Key{KeyBackspace}})
		if name != c.want || !utf8.ValidString(name) {
			t.Errorf("Backspace on %q gave %q (valid %v), want %q", c.in, name, utf8.ValidString(name), c.want)
		}
	}
}

// MaxLen counts runes, and Accept chooses the runes.
func TestFormTextFieldMaxLenAndAccept(t *testing.T) {
	name := ""
	f := NewForm(FormConfig{Fields: []FormField{{Label: "Name", Type: FieldText, Text: &name, MaxLen: 4,
		Accept: func(r rune) bool { return r != '/' }}}})
	f.Draw(noopRenderer{}, 800, 600, DefaultTheme())
	f.Update(Input{Chars: []rune{'a', '/', 'ñ', 'b', 'c', 'd', 'e'}})
	if name != "añbc" {
		t.Errorf("name = %q, want %q (no slash, four runes at most)", name, "añbc")
	}
	f.Update(Input{Chars: []rune{'x'}})
	if name != "añbc" {
		t.Errorf("a rune past MaxLen was let in: %q", name)
	}
	f.Update(Input{Keys: []Key{KeyBackspace}})
	f.Update(Input{Chars: []rune{'z'}})
	if name != "añbz" {
		t.Errorf("after Backspace and a rune: %q, want %q", name, "añbz")
	}
	// no MaxLen and no Accept: as before
	free := ""
	g := NewForm(FormConfig{Fields: []FormField{{Label: "Name", Type: FieldText, Text: &free}}})
	g.Draw(noopRenderer{}, 800, 600, DefaultTheme())
	g.Update(Input{Chars: []rune("a/b\x01c")})
	if free != "a/bc" {
		t.Errorf("a plain field holds %q, want %q", free, "a/bc")
	}
}
