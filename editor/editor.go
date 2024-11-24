package editor

import (
	"github.com/gdamore/tcell/v2"
)

const (
	PROMPT_SCREEN_LINE_BEGIN = 28
	PROMPT_SCREEN_COL_BEGIN  = 0

	LINE_CELL_ROW = 30
	LINE_CELL_COL = 30

	UPPER_CURSOR_BOUNDS  = 3
	BOTTOM_CURSOR_BOUNDS = 3
)
const (
	UP = iota
	DOWN
	RIGHT
	LEFT
)
const (
	INSERT_MODE = iota
	EXIT_MODE
	SEARCH_MODE
)

const (
	ON_WHICHTEXT = iota
	ON_NEWTEXT
	SEARCH_MODE_BUFFER_COUNT
)

const (
	SEARCH = iota
	REPLACE
)

type EditorConfiguration struct {
	Filepath *string
}

type EditorSearchModeParams struct {
	buffers       [SEARCH_MODE_BUFFER_COUNT]string
	locations     []Location
	current       int // points to the current location on which the cursor is focused
	whichMode     int
	currentBuffer int
	hasReplaced   bool
}

type Editor struct {
	screen          tcell.Screen
	buffer          Buffer
	realCursor      Location
	relativeCursor  Location
	renderingCursor Location
	config          EditorConfiguration
	mode            int
	searchParams    EditorSearchModeParams
}

// constructor for the editor structure
func New(editorConfig EditorConfiguration) (*Editor, error) {
	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}

	err = screen.Init()
	if err != nil {
		screen.Fini()
		return nil, err
	}

	editorStyle := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite)
	screen.SetStyle(editorStyle)

	return &Editor{
		screen:          screen,
		buffer:          NewBuffer(),
		realCursor:      Location{},
		relativeCursor:  Location{},
		renderingCursor: Location{},
		mode:            INSERT_MODE,
		config:          editorConfig,
		searchParams:    EditorSearchModeParams{},
	}, nil
}

// close the editor (remove the editor screen)
func (editor *Editor) Close() {
	editor.screen.Fini()
}

// return if the editor should still be running or not
func (editor *Editor) ShouldNotQuit() bool {
	return editor.mode != EXIT_MODE
}

// set the editor to the quitting mode
func (editor *Editor) Quit() {
	editor.mode = EXIT_MODE
}

// quit the editor and save into a hardcoded filepath
func (editor *Editor) quitAndSave() error {
	editor.Quit()
	return editor.save()
}
