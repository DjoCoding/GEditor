package editor

import (
	"github.com/gdamore/tcell/v2"
)

// well-defined consts
const (
	PROMPT_SCREEN_LINE_BEGIN = 28
	PROMPT_SCREEN_COL_BEGIN  = 0

	LINE_CELL_ROW = 30
	LINE_CELL_COL = 30

	UPPER_CURSOR_BOUNDS  = 3
	BOTTOM_CURSOR_BOUNDS = 3
)

// search mode sub modes
const (
	ON_SEARCH_ONLY = iota
	ON_SEARCH_AND_REPLACE
)

// modes
const (
	INSERT_MODE = iota
	EXIT_MODE
	SEARCH_MODE
)

// buffers
const (
	ON_WHICHTEXT = iota
	ON_NEWTEXT
	SEARCH_MODE_BUFFER_COUNT
)

// gestures
const (
	UP = iota
	DOWN
	RIGHT
	LEFT
)

type EditorMode int

type EditorConfiguration struct {
	Filepath *string
}

type EditorSearchModeParams struct {
	buffers   [SEARCH_MODE_BUFFER_COUNT]string
	locations []Location
	current   int // points to the current location on which the cursor is focused
}

type Editor struct {
	screen          tcell.Screen
	buffer          Buffer
	realCursor      Location
	relativeCursor  Location
	renderingCursor Location
	config          EditorConfiguration
	mode            EditorMode
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

// return if the editor should still be running or not
func (editor *Editor) ShouldNotQuit() bool {
	return editor.mode != EXIT_MODE
}

// set the editor to the quitting mode
func (editor *Editor) Quit() {
	editor.setMode(EXIT_MODE)
}

// close the editor (remove the editor screen)
func (editor *Editor) Close() {
	editor.screen.Fini()
}

// get the mode of the editor in a form of a string
func (editor *Editor) getModeAsString() string {
	switch editor.mode {
	case INSERT_MODE:
		return "INSERT"
	case SEARCH_MODE:
		return "SEARCH"
	case EXIT_MODE:
		return "EXIT"
	}

	return ""
}

func (editor *Editor) setMode(mode EditorMode) {
	editor.mode = mode
}
