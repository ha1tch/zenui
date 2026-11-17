package zenui

import "testing"

func TestMenuBarSetLabelAndLabelFor(t *testing.T) {
	mb := NewMenuBar(MenuBarConfig{Items: testBarItems()})
	first := mb.LabelFor(0)
	if first == "" {
		t.Fatal("the first item has no label")
	}
	mb.SetLabel(0, "SD Card")
	if got := mb.LabelFor(0); got != "SD Card" {
		t.Errorf("after SetLabel: %q", got)
	}
	// Out of range: nothing changes, nothing is returned.
	mb.SetLabel(-1, "x")
	mb.SetLabel(mb.LabelCount(), "x")
	if mb.LabelFor(-1) != "" || mb.LabelFor(mb.LabelCount()) != "" || mb.LabelFor(0) != "SD Card" {
		t.Error("an out-of-range index changed or returned a label")
	}
}
