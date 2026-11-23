package zenui

import (
	"reflect"
	"strings"
	"testing"
)

func TestLinkSpans(t *testing.T) {
	for _, c := range []struct {
		in   string
		bare bool
		want []textSpan
	}{
		{"no links here", true, []textSpan{{text: "no links here"}}},
		{"see [the site](https://a.example/x) now", false,
			[]textSpan{{text: "see "}, {text: "the site", url: "https://a.example/x"}, {text: " now"}}},
		{"  https://a.example/x", true,
			[]textSpan{{text: "  "}, {text: "https://a.example/x", url: "https://a.example/x"}}},
		{"at https://a.example/x.", true, // the full stop is the sentence's
			[]textSpan{{text: "at "}, {text: "https://a.example/x", url: "https://a.example/x"}, {text: "."}}},
		{"(https://a.example/x)", true, // so is the bracket
			[]textSpan{{text: "("}, {text: "https://a.example/x", url: "https://a.example/x"}, {text: ")"}}},
		{"https://a.example/f(x)", true, // a bracket with its opener in the address is the address's
			[]textSpan{{text: "https://a.example/f(x)", url: "https://a.example/f(x)"}}},
		{"([codeberg.org/x](https://codeberg.org/x))", false,
			[]textSpan{{text: "("}, {text: "codeberg.org/x", url: "https://codeberg.org/x"}, {text: ")"}}},
		{"a bare https://a.example/x is not one without bare", false,
			[]textSpan{{text: "a bare https://a.example/x is not one without bare"}}},
		{"xhttps://a.example", true, []textSpan{{text: "xhttps://a.example"}}}, // inside a word
		{"https://", true, []textSpan{{text: "https://"}}},                     // nothing after it
		// Only the web: no other scheme is a link, in either form.
		{"[run](file:///bin/sh) and ftp://a.example and javascript:alert(1)", true,
			[]textSpan{{text: "[run](file:///bin/sh) and ftp://a.example and javascript:alert(1)"}}},
		{"[x](https://a.example/with space)", false, []textSpan{{text: "[x](https://a.example/with space)"}}},
		{"[](https://a.example)", false, []textSpan{{text: "[](https://a.example)"}}},
		{"[unclosed](https://a.example", false, []textSpan{{text: "[unclosed](https://a.example"}}},
	} {
		if got := linkSpans(c.in, c.bare); !reflect.DeepEqual(got, c.want) {
			t.Errorf("linkSpans(%q, %v) = %q, want %q", c.in, c.bare, got, c.want)
		}
	}
}

// seen is what a person reads of a line: the links as their text.
func seen(line string) string {
	var b strings.Builder
	for _, sp := range linkSpans(line, false) {
		b.WriteString(sp.text)
	}
	return b.String()
}

func TestWrapLinkedTextKeepsEveryPartOfALinkWorking(t *testing.T) {
	const url = "https://example.org/a/very/long/address"
	in := "Thanks to [a person with a long name](" + url + ") for the loan of a machine, and see [short](https://b.example)."
	lines := WrapLinkedText(in, 24, "")
	if len(lines) < 4 {
		t.Fatalf("%d lines, want it wrapped over several: %q", len(lines), lines)
	}
	var all []string
	for _, l := range lines {
		if n := len(seen(l)); n > 24 {
			t.Errorf("line %q is %d columns seen, want at most 24", l, n)
		}
		all = append(all, seen(l))
	}
	if got, want := strings.Join(all, " "), seen(in); got != want {
		t.Errorf("the wrapped text reads %q, want %q", got, want)
	}
	// The link's words are on more than one line, and each line's part goes to the address.
	parts := 0
	for _, l := range lines {
		for _, sp := range linkSpans(l, false) {
			if sp.url == url {
				parts++
			}
		}
	}
	if parts < 2 {
		t.Errorf("the long link is in %d parts, want it carried over a break: %q", parts, lines)
	}
	if !strings.Contains(strings.Join(lines, "\n"), "[short](https://b.example)") {
		t.Errorf("the short link did not come through whole: %q", lines)
	}
}

func TestWrapLinkedTextPlainAndIndented(t *testing.T) {
	got := WrapLinkedText("one two three four five six", 9, "  ")
	want := []string{"  one two", "  three", "  four five", "  six"}
	if len(got) != 4 || got[0] != want[0] || got[1] != want[1] || got[2] != want[2] || got[3] != want[3] {
		// "three four" is ten columns: it does not fit in nine.
		t.Errorf("got %q, want %q", got, want)
	}
	if got := WrapLinkedText("", 20, ""); len(got) != 1 || got[0] != "" {
		t.Errorf("empty text wrapped to %q, want one empty line", got)
	}
	if got := WrapLinkedText("unbreakable_long_word", 5, ""); len(got) != 1 || got[0] != "unbreakable_long_word" {
		t.Errorf("a word longer than the width wrapped to %q, want it whole on one line", got)
	}
}

const linksText = "# Title\n\nbody\n  https://a.example/x\nsee [the page](https://b.example/y) here"

// click returns an Input pressing the left button at the middle of r.
func clickAt(r Rect) Input {
	return Input{MouseX: r.X + r.W/2, MouseY: r.Y + r.H/2, MousePressed: true, MouseDown: true}
}

func TestTextModalLinksOpenWhenClicked(t *testing.T) {
	theme := DarkTheme()
	var opened []string
	m := NewTextModal("ABOUT", linksText, true).WithLinks(func(u string) { opened = append(opened, u) })
	r := newDrawRecorder()
	m.Draw(r, 1200, 900, theme)
	if len(m.links) != 2 {
		t.Fatalf("%d links drawn, want 2: %+v", len(m.links), m.links)
	}
	// Each link is drawn in the accent colour and underlined.
	for _, l := range m.links {
		underlined := false
		for _, c := range *r.calls {
			if c.kind == "fill" && c.rect.X == l.r.X && c.rect.W == l.r.W && c.rect.H == 1 && c.rect.Y >= l.r.Y && c.rect.Y < l.r.Y+l.r.H && c.col == theme.DirText {
				underlined = true
			}
		}
		if !underlined {
			t.Errorf("the link to %s is not underlined", l.url)
		}
	}
	for i, want := range []string{"https://a.example/x", "https://b.example/y"} {
		if !m.Update(clickAt(m.links[i].r)) {
			t.Errorf("a click on a link closed the modal")
		}
		if len(opened) != i+1 || opened[i] != want {
			t.Fatalf("opened %q after click %d, want %q last", opened, i+1, want)
		}
	}
	// Text that is not a link, and the panel round it, do nothing.
	before := len(opened)
	if !m.Update(Input{MouseX: m.body.X + 2, MouseY: m.body.Y + 2, MousePressed: true}) || len(opened) != before {
		t.Errorf("a click on a heading opened %q or closed the modal", opened[before:])
	}
	// A click outside the panel still closes it.
	if m.Update(Input{MouseX: 2, MouseY: 2, MousePressed: true}) {
		t.Error("a click outside the panel did not close the modal")
	}
	if len(opened) != before {
		t.Errorf("a click outside opened %q", opened[before:])
	}
}

func TestTextModalOverLinkFollowsThePointer(t *testing.T) {
	m := NewTextModal("ABOUT", linksText, true).WithLinks(func(string) {})
	m.Draw(newDrawRecorder(), 1200, 900, DarkTheme())
	if m.OverLink() {
		t.Error("over a link before the pointer has moved")
	}
	in := Input{MouseX: m.links[0].r.X + 3, MouseY: m.links[0].r.Y + 3}
	m.Update(in)
	if !m.OverLink() {
		t.Error("not over the link the pointer is on")
	}
	m.Update(Input{MouseX: m.body.X + 2, MouseY: m.body.Y + 2})
	if m.OverLink() {
		t.Error("over a link with the pointer on a heading")
	}
}

// A modal made without WithLinks reads the same text as plain text: the text of an
// info window comes from a file, and nothing in it may be a link.
func TestTextModalWithoutLinksHasNone(t *testing.T) {
	theme := DarkTheme()
	m := NewTextModal("INFO", linksText, true)
	r := newDrawRecorder()
	m.Draw(r, 1200, 900, theme)
	if len(m.links) != 0 || m.OverLink() {
		t.Errorf("%d links in a modal that did not ask for them", len(m.links))
	}
	for _, c := range *r.calls {
		if c.kind == "text" && strings.Contains(c.label, "b.example") && c.label != "see [the page](https://b.example/y) here" {
			t.Errorf("the line was split into pieces without links: %q", c.label)
		}
	}
	for i := range m.lines {
		m.Update(Input{MouseX: m.body.X + 20, MouseY: m.body.Y + i*m.bodyLH + 2, MousePressed: true})
	}
}

func TestTextModalLinksScrollWithTheText(t *testing.T) {
	var opened []string
	lines := strings.Repeat("filler\n", 60) + "[the end](https://c.example/z)"
	m := NewTextModal("HELP", lines, false).WithLinks(func(u string) { opened = append(opened, u) })
	m.Draw(newDrawRecorder(), 1200, 900, DarkTheme())
	if len(m.links) != 0 {
		t.Fatalf("a link off the bottom of the window is live: %+v", m.links)
	}
	m.Update(Input{Keys: []Key{KeyEnd}})
	m.Draw(newDrawRecorder(), 1200, 900, DarkTheme())
	if len(m.links) != 1 {
		t.Fatalf("%d links after scrolling to the end, want 1", len(m.links))
	}
	m.Update(clickAt(m.links[0].r))
	if len(opened) != 1 || opened[0] != "https://c.example/z" {
		t.Errorf("opened %q, want the link at the end", opened)
	}
}
