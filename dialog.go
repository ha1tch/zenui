package zenui

// Mode selects the dialog's behaviour.
type Mode int

const (
	// ModeOpen picks an existing file; the action button is disabled unless a
	// file (not a directory) is selected.
	ModeOpen Mode = iota
	// ModeSave names a file to write; the action button is enabled whenever the
	// filename field is non-empty. An existing name may be picked to overwrite.
	ModeSave
)

// EntryPreview is a small rendered thumbnail for an entry, supplied by the host
// via DialogConfig.Preview. Pixels is row-major, length W*H; a fully transparent
// pixel (A==0) is not drawn (the panel background shows through).
type EntryPreview struct {
	W, H   int
	Pixels []Colour
}

// DialogConfig sets up a dialog. Only FS is required; the rest have sensible
// defaults.
type DialogConfig struct {
	Mode       Mode
	Title      string   // window title; defaults to "Open"/"Save"
	StartDir   string   // initial directory; defaults to FS.UserHome()
	Filters    []string // extensions to show (e.g. {"zcut","scr"}); empty = all
	DefaultExt string   // appended on save if the typed name has no extension
	FS         FS       // filesystem; defaults to OSFS
	// Preview, if set, is called for the selected entry to obtain a thumbnail for
	// the preview pane. container is the path of the container the entry lives in
	// (empty for a plain directory entry). Returning nil means "no thumbnail"
	// (text metadata, if any, is still shown).
	Preview func(container string, entry Entry) *EntryPreview
	// StartContainer, if set, opens the dialog directly inside this container file
	// (e.g. a dropped .zbun), listing its entries; ".." then lands in the
	// container's directory.
	StartContainer string

	// Name is the text the name field starts with (a Save's suggested file name,
	// or an Open's typed path).
	Name string
	// FilterSets, if set, gives the person a choice of file types: a selector
	// beside the name field, and Ctrl+Tab to go to the next. The dialog adds a
	// last set, "All files", unless NoAllFiles. Filters is the shorthand for a
	// single set with no choice, and is ignored when FilterSets is set.
	FilterSets []FilterSet
	NoAllFiles bool
	// View is how the person last looked at the folder (hidden files, details,
	// order, the chosen file type); read it back with Dialog.View to keep it.
	View View
	// Recent lists full paths, the most recent first. When it is not empty the
	// sidebar has a first entry for it, listing the files that pass the chosen
	// file type; opening one checks that it is still there.
	Recent []string
	// Places, if set, replaces the sidebar's favourites (the filesystem's own);
	// ExtraPlaces adds to them. A Place is a label and a folder.
	Places      []Place
	ExtraPlaces []Place
	// Validate, if set, is asked about the path the person has chosen, in Open
	// and in Save, before the dialog finishes. A non-empty answer is shown as the
	// error and the dialog stays open.
	Validate func(path string) string
	// NoOverwriteConfirm makes a Save replace an existing file without asking.
	// By default the dialog asks once: the action button reads REPLACE and Enter
	// (or the button) confirms, Escape keeps the file.
	NoOverwriteConfirm bool
	// Labels are the words the dialog draws, for another language or style.
	Labels DialogLabels
	// Scale is the integer text scale (default 2). Width and Height are the
	// largest the panel grows to, in pixels (default 760 x 560).
	Scale         int
	Width, Height int
	// DoubleClickSeconds is the longest time between the two clicks of a double
	// click that opens an entry. Zero means half a second; a negative value
	// means no limit (a second click on the selected row always opens it).
	// The time is counted from Input.DeltaTime, so a host that does not set it
	// gets the "no limit" behaviour whatever this says.
	DoubleClickSeconds float64
}

// Status is the dialog's lifecycle state.
type Status int

const (
	// Active: the dialog is open and should keep being drawn.
	Active Status = iota
	// Accepted: the user confirmed a choice; read Result for the path.
	Accepted
	// Cancelled: the user dismissed the dialog.
	Cancelled
	// Toggled: a checkbox-style Menu item (Item.Toggle) was clicked --
	// its Checked value already flipped, and the menu stays Active
	// rather than closing. Read Result for which item. Never returned
	// by Dialog, only by Menu.
	Toggled
)

// Dialog is a file Open/Save dialog. Construct with New, then each frame call
// Update(input) and Draw(renderer). When Update returns a status other than
// Active the dialog is finished and Result holds the chosen path (for Accepted).
type Dialog struct {
	cfg DialogConfig
	fs  FS

	dir       string  // current directory (absolute)
	container string  // if non-empty, we are browsing inside this container file
	entries   []Entry // filtered, sorted listing of dir (or container)
	sel       int     // index into entries, or -1
	scroll    int     // first visible row
	name      string  // filename field contents (Save mode, and echo of selection)
	err       string  // last filesystem error, shown in the status line
	places    []Place // favourites shown in the sidebar

	status Status
	result string

	// layout caches from the last Draw, used by Update's hit-testing.
	bounds      Rect
	sideRect    Rect
	placeRects  []Rect
	listRect    Rect
	previewRect Rect
	rowH        int
	nameRect    Rect
	okRect      Rect
	cancRect    Rect
	upRect      Rect

	raw       []Entry // the listing before the view's hidden, file-type and order are applied
	recent    bool    // the list on show is the recent files
	typed     bool    // the name field holds the person's own text, not the echo of the selection
	ed        lineEdit
	hint      string // a note in the status line (matches of a completion, and so on)
	confirm   string // the path a Save is asking to replace
	sets      []FilterSet
	view      View
	custom    string // a pattern typed in the name field (*.tap) that limits the list
	popup     bool   // the file-type list is open
	popupSel  int
	infoChars int // the longest Entry.Detail on show, in runes
	clock     float64
	lastClick float64
	lastIdx   int
	textW     func(string) int

	filterRect  Rect
	headRect    Rect
	hiddenRect  Rect
	detailsRect Rect
	statusRect  Rect
	popupRect   Rect
	popupRows   []Rect
	cols        []dlgCol
}

// NewDialog creates a dialog from cfg, applying defaults and loading the
// start directory. It never returns nil; a bad start directory leaves the
// dialog at the filesystem root-ish home with an error message visible.
func NewDialog(cfg DialogConfig) *Dialog {
	if cfg.FS == nil {
		cfg.FS = OSFS{}
	}
	if cfg.Title == "" {
		if cfg.Mode == ModeSave {
			cfg.Title = "Save"
		} else {
			cfg.Title = "Open"
		}
	}
	d := &Dialog{cfg: cfg, fs: cfg.FS, sel: -1, lastIdx: -1}
	d.sets = buildSets(cfg)
	d.view = cfg.View
	if len(d.sets) > 0 {
		d.view.Filter = clampInt(d.view.Filter, 0, len(d.sets)-1)
	} else {
		d.view.Filter = 0
	}
	if !d.hiddenSupported() {
		d.view.ShowHidden = false
	}
	if cfg.Places != nil {
		d.places = append(d.places, cfg.Places...)
	} else {
		d.places = d.fs.Places()
	}
	d.places = append(d.places, cfg.ExtraPlaces...)
	if len(cfg.Recent) > 0 {
		d.places = append([]Place{{Label: pick(cfg.Labels.Recent, "Recent"), Path: recentPlace}}, d.places...)
	}
	start := cfg.StartDir
	if start == "" {
		start = d.fs.UserHome()
	}
	if abs, err := d.fs.Abs(start); err == nil {
		start = abs
	}
	d.setDir(start)
	// Optionally descend straight into a container (e.g. a dropped bundle).
	if cfg.StartContainer != "" {
		if abs, err := d.fs.Abs(cfg.StartContainer); err == nil {
			d.dir = d.fs.Dir(abs)
			d.setContainer(abs)
		}
	}
	if cfg.Name != "" {
		d.setName(cfg.Name, true)
	}
	return d
}

// Result returns the chosen path (valid once Status() == Accepted).
func (d *Dialog) Result() string { return d.result }

// Status returns the dialog's current lifecycle state.
func (d *Dialog) Status() Status { return d.status }

// Dir returns the directory currently being browsed.
func (d *Dialog) Dir() string { return d.dir }
