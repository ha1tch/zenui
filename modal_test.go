package zenui

import "testing"

func TestNewModalNeverNil(t *testing.T) {
	m := NewModal(ModalConfig{
		Form:    FormConfig{Fields: []FormField{{Label: "Are you sure?", Type: FieldLabel}}},
		Buttons: []string{"OK"},
	})
	if m == nil {
		t.Fatal("NewModal returned nil")
	}
	if m.Status() != Active {
		t.Errorf("Status() = %v, want Active for a freshly constructed modal", m.Status())
	}
}

func TestModalButtonsPositionedLeftToRight(t *testing.T) {
	m := NewModal(ModalConfig{
		Form:    FormConfig{Fields: []FormField{{Label: "Msg", Type: FieldLabel}}},
		Buttons: []string{"Cancel", "OK"},
	})
	m.Draw(noopRenderer{}, 800, 600, DefaultTheme())

	if len(m.buttonRects) != 2 {
		t.Fatalf("len(buttonRects) = %d, want 2", len(m.buttonRects))
	}
	cancel, ok := m.buttonRects[0], m.buttonRects[1]
	if ok.X < cancel.X+cancel.W {
		t.Errorf("\"OK\" (X=%d) should be to the right of \"Cancel\" (X+W=%d) -- Buttons order is left to right", ok.X, cancel.X+cancel.W)
	}
}

func TestModalButtonsAreInsideTheWindowsFrame(t *testing.T) {
	for _, buttons := range [][]string{{"OK"}, {"Cancel", "8 x 8", "16 x 16"}} {
		m := NewModal(ModalConfig{
			Form:    FormConfig{Title: "New project", Fields: []FormField{{Label: "Msg", Type: FieldLabel}}},
			Buttons: buttons,
		})
		m.Draw(noopRenderer{}, 800, 600, DefaultTheme())

		frame := m.form.Bounds()
		if frame != m.bounds {
			t.Errorf("%v: the window is %+v but the form's frame is %+v, want one frame", buttons, m.bounds, frame)
		}
		if frame.X*2+frame.W != 800 || frame.Y*2+frame.H != 600 {
			t.Errorf("%v: the window %+v is not centred on an 800 by 600 screen", buttons, frame)
		}
		field := m.form.fieldRects[0]
		const pad = 4 // half the default grid
		for i, b := range m.buttonRects {
			if b.X < frame.X+pad || b.Y < frame.Y || b.X+b.W > frame.X+frame.W-pad || b.Y+b.H > frame.Y+frame.H-pad {
				t.Errorf("%v: button %d %+v is not inside the frame %+v with its margin", buttons, i, b, frame)
			}
			if b.Y < field.Y+field.H {
				t.Errorf("%v: button %d (Y=%d) is not beneath the message (bottom %d)", buttons, i, b.Y, field.Y+field.H)
			}
		}
		last := m.buttonRects[len(m.buttonRects)-1]
		if got := frame.X + frame.W - (last.X + last.W); got != pad {
			t.Errorf("%v: the last button is %d from the frame's right edge, want %d", buttons, got, pad)
		}
	}
}

func TestModalClickingButtonAccepts(t *testing.T) {
	m := NewModal(ModalConfig{
		Form:              FormConfig{Fields: []FormField{{Label: "Msg", Type: FieldLabel}}},
		Buttons:           []string{"Cancel", "OK"},
		CancelButtonIndex: 0,
	})
	m.Draw(noopRenderer{}, 800, 600, DefaultTheme())

	okRect := m.buttonRects[1]
	status := m.Update(Input{MouseX: okRect.X + 1, MouseY: okRect.Y + 1, MousePressed: true})

	if status != Accepted {
		t.Errorf("Update() = %v, want Accepted", status)
	}
	if m.Result() != 1 {
		t.Errorf("Result() = %d, want 1 (\"OK\", the second button)", m.Result())
	}
}

func TestModalClickingCancelButtonCancels(t *testing.T) {
	m := NewModal(ModalConfig{
		Form:              FormConfig{Fields: []FormField{{Label: "Msg", Type: FieldLabel}}},
		Buttons:           []string{"Cancel", "OK"},
		CancelButtonIndex: 0,
	})
	m.Draw(noopRenderer{}, 800, 600, DefaultTheme())

	cancelRect := m.buttonRects[0]
	status := m.Update(Input{MouseX: cancelRect.X + 1, MouseY: cancelRect.Y + 1, MousePressed: true})

	if status != Cancelled {
		t.Errorf("Update() = %v, want Cancelled", status)
	}
	if m.Result() != 0 {
		t.Errorf("Result() = %d, want 0 (\"Cancel\", the CancelButtonIndex)", m.Result())
	}
}

func TestModalEscapeTriggersCancelButtonIndex(t *testing.T) {
	m := NewModal(ModalConfig{
		Form:              FormConfig{Fields: []FormField{{Label: "Msg", Type: FieldLabel}}},
		Buttons:           []string{"No", "Yes"},
		CancelButtonIndex: 0,
	})
	m.Draw(noopRenderer{}, 800, 600, DefaultTheme())

	status := m.Update(Input{Keys: []Key{KeyEscape}})
	if status != Cancelled {
		t.Errorf("Update() after Escape = %v, want Cancelled", status)
	}
	if m.Result() != 0 {
		t.Errorf("Result() = %d, want 0 (CancelButtonIndex)", m.Result())
	}
}

func TestModalEscapeDoesNothingWithoutCancelButtonIndex(t *testing.T) {
	m := NewModal(ModalConfig{
		Form:              FormConfig{Fields: []FormField{{Label: "Msg", Type: FieldLabel}}},
		Buttons:           []string{"OK"},
		CancelButtonIndex: -1, // explicit: no Escape handling
	})
	m.Draw(noopRenderer{}, 800, 600, DefaultTheme())

	status := m.Update(Input{Keys: []Key{KeyEscape}})
	if status != Active {
		t.Errorf("Update() after Escape with CancelButtonIndex=-1 = %v, want Active (unaffected)", status)
	}
}

func TestModalFormFieldsWorkThroughModal(t *testing.T) {
	name := ""
	m := NewModal(ModalConfig{
		Form:    FormConfig{Fields: []FormField{{Label: "Name", Type: FieldText, Text: &name}}},
		Buttons: []string{"OK"},
	})
	m.Draw(noopRenderer{}, 800, 600, DefaultTheme())

	m.Update(Input{Chars: []rune{'H', 'i'}})
	if name != "Hi" {
		t.Errorf("name = %q, want %q (Modal.Update should delegate typing to the wrapped Form)", name, "Hi")
	}
}

func TestModalDrawNoopsOnceClosed(t *testing.T) {
	r := newDrawRecorder()
	m := NewModal(ModalConfig{
		Form:    FormConfig{Fields: []FormField{{Label: "Msg", Type: FieldLabel}}},
		Buttons: []string{"OK"},
	})
	m.Draw(r, 800, 600, DefaultTheme())
	okRect := m.buttonRects[0]
	m.Update(Input{MouseX: okRect.X + 1, MouseY: okRect.Y + 1, MousePressed: true})

	r2 := newDrawRecorder()
	m.Draw(r2, 800, 600, DefaultTheme())
	if len(*r2.calls) != 0 {
		t.Errorf("Draw after Accepted made %d calls, want 0 (a closed modal should draw nothing)", len(*r2.calls))
	}
}
