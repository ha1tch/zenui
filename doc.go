// Package zenui is a renderer-agnostic widget toolkit: a file Open/Save dialog,
// menus and a menu bar, forms, message boxes, option panels, tool and palette
// choosers and a preview pane.
//
// It owns the logic of the widgets (the file dialog's directory navigation,
// listing, filtering, selection, scrolling, filename entry and keyboard
// handling, for one) but draws nothing itself. The host supplies a small
// Renderer (rectangles, gradients, lines, text, clipping) and an Input snapshot
// (mouse and keyboard) each frame.
// Because the package imports no graphics library, the same widgets run under
// raylib, a software framebuffer, or a headless test harness.
//
// It is shared by zenimate, zenzx and zenplay. Apart from its zxpalette
// subpackage it has no dependencies outside the standard library.
package zenui
