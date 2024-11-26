package editor

import (
	"os"

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
	SELECTION_MODE
	NAVIGATION_MODE
)

const (
	INPUT_TEXT = iota
	NEW_TEXT
	EDITOR_BUFFER_COUNT
)

const (
	SEARCH = iota
	REPLACE
)

type EditorConfiguration struct {
	OpenedFile  string // can be a dir
	CurrentFile string // current handled file
}

type EditorSelectionModeParams struct {
	startLocation Location
	endLocation   Location
}

type EditorSearchModeParams struct {
	locations   []Location
	current     int // points to the current location on which the cursor is focused
	whichMode   int
	hasReplaced bool
}

type EditorNavigationModeParams struct {
	files            []os.DirEntry
	currentFileIndex int
}

type EditorInternalInput struct {
	buffers [EDITOR_BUFFER_COUNT]string
	enabled bool
	current int
	req     string
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
	selParams       EditorSelectionModeParams
	navParams       EditorNavigationModeParams
	input           EditorInternalInput
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
		buffer:          newBuffer(),
		realCursor:      Location{},
		relativeCursor:  Location{},
		renderingCursor: Location{},
		mode:            INSERT_MODE,
		config:          editorConfig,
		searchParams:    EditorSearchModeParams{},
		selParams:       EditorSelectionModeParams{},
		input:           EditorInternalInput{},
		navParams:       EditorNavigationModeParams{},
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
func (editor *Editor) saveAndquit() error {
	editor.Quit()

	if editor.config.CurrentFile != "" {
		return editor.save()
	}

	return nil
}
