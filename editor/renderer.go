package editor

import (
	"math"
	"strconv"

	"github.com/gdamore/tcell/v2"
)

func (editor *Editor) renderLineInNormalMode(lineIndex int, row int) {
	line := editor.buffer.lines[lineIndex]

	for i, c := range line.getContent() {
		editor.screen.SetContent(i, row, c, nil, tcell.StyleDefault)
	}
}

func (editor *Editor) renderLineInSearchMode(lineIndex int, row int) {
	style := tcell.StyleDefault.Bold(true).Underline(true).Background(tcell.ColorDarkCyan)
	line := editor.buffer.lines[lineIndex]

	count := 0

	for i, c := range line.getContent() {
		currentLocation := newLocation(lineIndex, i)
		found := editor.lookupLocationInSearchLocations(currentLocation)

		if found {
			count = len(editor.input.buffers[INPUT_TEXT])
		}

		if count > 0 {
			editor.screen.SetContent(i, row, c, nil, style)
			count--
			continue
		}

		editor.screen.SetContent(i, row, c, nil, tcell.StyleDefault)
	}
}

func (editor *Editor) renderLineInSelectionMode(lineIndex int, row int) {
	style := tcell.StyleDefault.Background(tcell.ColorBlue)
	line := editor.buffer.lines[lineIndex]

	for i, c := range line.getContent() {
		currentLocation := newLocation(lineIndex, i)
		if editor.checkLocationInSelectionModeBounds(currentLocation) {
			editor.screen.SetContent(i, row, c, nil, style)
			continue
		}

		editor.screen.SetContent(i, row, c, nil, tcell.StyleDefault)
	}
}

func (editor *Editor) updateRenderingCursor() {
	for editor.realCursor.getLine() < editor.renderingCursor.getLine()+UPPER_CURSOR_BOUNDS {
		editor.renderingCursor.setLine(editor.renderingCursor.getLine() - 1)
		if editor.renderingCursor.getLine() < 0 {
			editor.renderingCursor.setLine(0)
			return
		}
	}

	_, h := editor.screen.Size()
	h -= BOTTOM_CURSOR_BOUNDS
	for editor.realCursor.getLine() > editor.renderingCursor.getLine()+h-BOTTOM_CURSOR_BOUNDS {
		editor.renderingCursor.setLine(editor.renderingCursor.getLine() + 1)
	}
}

func (editor *Editor) updateRelativeCursor() {
	editor.relativeCursor.set(editor.realCursor.getLine()-editor.renderingCursor.getLine(), editor.realCursor.getCol()-editor.renderingCursor.getCol())
}

func (editor *Editor) getNumberLinesToRender() int {
	_, h := editor.screen.Size()
	h -= BOTTOM_CURSOR_BOUNDS

	return int(math.Min(float64(h), float64(editor.buffer.count()-editor.renderingCursor.getLine())))
}

// render the content of the editor buffer in the normal mode
func (editor *Editor) renderContentInNormalMode() {
	numberLinesToRender := editor.getNumberLinesToRender()
	for i := 0; i < numberLinesToRender; i++ {
		editor.renderLineInNormalMode(editor.renderingCursor.getLine()+i, i)
	}
}

// render the content of the editor buffer in the search mode
func (editor *Editor) renderContentInSearchMode() {
	numberLinesToRender := editor.getNumberLinesToRender()
	for i := 0; i < numberLinesToRender; i++ {
		editor.renderLineInSearchMode(editor.renderingCursor.getLine()+i, i)
	}
}

func (editor *Editor) renderContentInSelectionMode() {
	numberLinesToRender := editor.getNumberLinesToRender()
	for i := 0; i < numberLinesToRender; i++ {
		editor.renderLineInSelectionMode(editor.renderingCursor.getLine()+i, i)
	}
}

// render the content of the editor buffer
func (editor *Editor) renderContent() {
	editor.updateRenderingCursor()

	switch editor.mode {
	case INSERT_MODE:
		editor.renderContentInNormalMode()
	case SEARCH_MODE:
		editor.renderContentInSearchMode()
	case SELECTION_MODE:
		editor.renderContentInSelectionMode()
	}
}

// render the cursor of the editor (real Cursor)
func (editor *Editor) renderCursor() {
	editor.updateRelativeCursor()
	editor.screen.ShowCursor(editor.relativeCursor.getCol(), editor.relativeCursor.getLine())
}

// render any text to the editor screen (helper function)
func (editor *Editor) renderText(line, col int, text string) {
	for i, c := range text {
		editor.screen.SetContent(col+i, line, c, nil, tcell.StyleDefault)
	}
}

// render information (mode, cursor)
func (editor *Editor) renderInfo() {
	lineString := strconv.Itoa(editor.realCursor.getLine())
	colString := strconv.Itoa(editor.realCursor.getCol())

	editor.renderText(LINE_CELL_ROW, LINE_CELL_COL, lineString)
	editor.renderText(LINE_CELL_ROW, LINE_CELL_COL+len(lineString), ":")
	editor.renderText(LINE_CELL_ROW, LINE_CELL_COL+len(lineString)+1, colString)

	if editor.inputBufferIsEnabled() {
		textToRender := editor.input.req + editor.input.buffers[editor.getInputCurrentBuffer()]
		editor.renderText(PROMPT_SCREEN_LINE_BEGIN+1, PROMPT_SCREEN_COL_BEGIN, textToRender)
	}
}

// render the content of the editor together with some information
func (editor *Editor) Render() {
	editor.screen.Clear()
	editor.renderContent()
	editor.renderInfo()
	editor.renderCursor()
	editor.screen.Show()
}
