package zenui

import (
	"strings"
	"testing"
)

func TestDrawProgressBar(t *testing.T) {
	theme := DarkTheme()
	rect := Rect{X: 10, Y: 20, W: 200, H: 12}
	fills := func(done, total int64) []drawCall {
		r := newDrawRecorder()
		DrawProgressBar(r, rect, done, total, theme)
		var out []drawCall
		for _, c := range *r.calls {
			if c.kind == "fill" {
				out = append(out, c)
			}
		}
		if last := (*r.calls)[len(*r.calls)-1]; last.kind != "stroke" || last.rect != rect || last.col != theme.Border {
			t.Errorf("%d of %d: the border is not drawn last, round the bar: %+v", done, total, last)
		}
		return out
	}
	if f := fills(30, 120); len(f) != 2 || f[0].rect != rect || f[0].col != theme.Field ||
		f[1].rect != (Rect{X: 10, Y: 20, W: 50, H: 12}) || f[1].col != theme.SelFill {
		t.Errorf("30 of 120: %+v", f)
	}
	if f := fills(0, 0); len(f) != 1 {
		t.Errorf("a total not known yet: %+v", f)
	}
	if f := fills(500, 120); len(f) != 2 || f[1].rect.W != 200 {
		t.Errorf("done past the total: %+v", f)
	}
}

func TestProgressText(t *testing.T) {
	if line := ProgressText(30, 120); !strings.HasPrefix(line, " 25%") {
		t.Errorf("30 of 120: %q", line)
	}
	if line := ProgressText(3<<20, 0); line != "3.0 MB" {
		t.Errorf("3 MB of an unknown total: %q", line)
	}
	if line := ProgressText(0, 0); line != "" {
		t.Errorf("nothing yet: %q", line)
	}
}

func TestWrapTextKeepsToTheWidth(t *testing.T) {
	path := "/home/someone/.config/zenzx/next/and/a/rather/long/way/further/down/the/tree"
	lines := WrapText(path, 30, "  ")
	if len(lines) < 3 {
		t.Fatalf("a %d character path came out as %d lines", len(path), len(lines))
	}
	joined := ""
	for _, l := range lines {
		if len(l) > 32 {
			t.Errorf("line %q is longer than the width and its indent", l)
		}
		joined += strings.TrimPrefix(l, "  ")
	}
	if joined != path {
		t.Fatalf("wrapping changed the text: %q", joined)
	}
	if got := WrapText("short", 30, "> "); len(got) != 1 || got[0] != "> short" {
		t.Errorf("a short line: %q", got)
	}
}

func TestFitTail(t *testing.T) {
	r := noopRenderer{} // 8 pixels a character at scale 1
	if got := FitTail(r, "https://example.org/a/long/address", 1, 80); got != "ng/address" {
		t.Errorf("%q", got)
	}
	if got := FitTail(r, "short", 1, 80); got != "short" {
		t.Errorf("%q", got)
	}
}
