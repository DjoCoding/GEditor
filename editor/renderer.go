package editor

import (
	"math"
	"strconv"

	"github.com/gdamore/tcell/v2"
)

func (e *Editor) renderLineInInsertMode(lineIndex int, row int) {
	line := e.buffer.lines[lineIndex]

	for i, c := range line.getContent() {
		e.screen.SetContent(i, row, c, nil, tcell.StyleDefault)
	}
}

func (e *Editor) renderLineInSearchMode(lineIndex int, row int) {
	style := tcell.StyleDefault.Bold(true).Underline(true).Background(tcell.ColorDarkCyan)
	line := e.buffer.lines[lineIndex]

	count := 0

	for i, c := range line.getContent() {
		currentLocation := newLocation(lineIndex, i)
		found := e.lookupLocationInSearchLocations(currentLocation)

		if found {
			count = len(e.input.buffers[INPUT_TEXT])
		}

		if count > 0 {
			e.screen.SetContent(i, row, c, nil, style)
			count--
			continue
		}

		e.screen.SetContent(i, row, c, nil, tcell.StyleDefault)
	}
}

func (e *Editor) renderLineInSelectionMode(lineIndex int, row int) {
	style := tcell.StyleDefault.Background(tcell.ColorBlue)
	line := e.buffer.lines[lineIndex]

	for i, c := range line.getContent() {
		currentLocation := newLocation(lineIndex, i)
		if e.checkLocationInSelectionModeBounds(currentLocation) {
			e.screen.SetContent(i, row, c, nil, style)
			continue
		}

		e.screen.SetContent(i, row, c, nil, tcell.StyleDefault)
	}
}

func (e *Editor) updateRenderingCursor() {
	for e.realCursor.getLine() < e.renderingCursor.getLine()+UPPER_CURSOR_BOUNDS {
		e.renderingCursor.setLine(e.renderingCursor.getLine() - 1)
		if e.renderingCursor.getLine() < 0 {
			e.renderingCursor.setLine(0)
			return
		}
	}

	_, h := e.screen.Size()
	h -= BOTTOM_CURSOR_BOUNDS
	for e.realCursor.getLine() > e.renderingCursor.getLine()+h-BOTTOM_CURSOR_BOUNDS {
		e.renderingCursor.setLine(e.renderingCursor.getLine() + 1)
	}
}

func (e *Editor) updateRelativeCursor() {
	e.relativeCursor.set(e.realCursor.getLine()-e.renderingCursor.getLine(), e.realCursor.getCol()-e.renderingCursor.getCol())
}

func (e *Editor) getNumberLinesToRender() int {
	_, h := e.screen.Size()
	h -= BOTTOM_CURSOR_BOUNDS

	return int(math.Min(float64(h), float64(e.buffer.count()-e.renderingCursor.getLine())))
}

// render the content of the e buffer in the normal mode
func (e *Editor) renderContentInInsertMode() {
	numberLinesToRender := e.getNumberLinesToRender()
	for i := 0; i < numberLinesToRender; i++ {
		e.renderLineInInsertMode(e.renderingCursor.getLine()+i, i)
	}
}

// render the content of the e buffer in the search mode
func (e *Editor) renderContentInSearchMode() {
	numberLinesToRender := e.getNumberLinesToRender()
	for i := 0; i < numberLinesToRender; i++ {
		e.renderLineInSearchMode(e.renderingCursor.getLine()+i, i)
	}
}

func (e *Editor) renderContentInSelectionMode() {
	numberLinesToRender := e.getNumberLinesToRender()
	for i := 0; i < numberLinesToRender; i++ {
		e.renderLineInSelectionMode(e.renderingCursor.getLine()+i, i)
	}
}

// render the content of the e buffer
func (e *Editor) renderContent() {
	e.updateRenderingCursor()

	switch e.mode {
	case INSERT_MODE:
		e.renderContentInInsertMode()
	case SEARCH_MODE:
		e.renderContentInSearchMode()
	case SELECTION_MODE:
		e.renderContentInSelectionMode()
	}
}

// render the cursor of the e (real Cursor)
func (e *Editor) renderCursor() {
	e.updateRelativeCursor()
	e.screen.ShowCursor(e.relativeCursor.getCol(), e.relativeCursor.getLine())
}

func (e *Editor) renderTextOnStyle(line, col int, text string, style tcell.Style) {
	for i, c := range text {
		e.screen.SetContent(col+i, line, c, nil, style)
	}
}

// render any text to the e screen (helper function)
func (e *Editor) renderText(line, col int, text string) {
	e.renderTextOnStyle(line, col, text, tcell.StyleDefault)
}

// render information (mode, cursor)
func (e *Editor) renderInfo() {
	lineString := strconv.Itoa(e.realCursor.getLine())
	colString := strconv.Itoa(e.realCursor.getCol())

	e.renderText(LINE_CELL_ROW, LINE_CELL_COL, lineString)
	e.renderText(LINE_CELL_ROW, LINE_CELL_COL+len(lineString), ":")
	e.renderText(LINE_CELL_ROW, LINE_CELL_COL+len(lineString)+1, colString)

	if e.inputBufferIsEnabled() {
		textToRender := e.input.req + e.input.buffers[e.getInputCurrentBuffer()]
		e.renderText(PROMPT_SCREEN_LINE_BEGIN+1, PROMPT_SCREEN_COL_BEGIN, textToRender)
	}
}

func (e *Editor) renderNavigation() {
	for i, file := range e.navParams.files {
		style := tcell.StyleDefault

		if i == e.navParams.currentFileIndex {
			style = style.Background(tcell.ColorGray)
			if file.IsDir() {
				style = style.Background(tcell.ColorDarkCyan)
			}
		}

		e.renderTextOnStyle(i, 0, e.config.OpenedFile+"/"+file.Name(), style)
	}
}

func (e *Editor) renderEditorTextOnScreen() {
	e.renderInfo()

	if e.mode == NAVIGATION_MODE {
		e.renderNavigation()
		return
	}

	e.renderContent()
	e.renderCursor()
}

// render the content of the editor together with some information
func (e *Editor) Render() {
	e.screen.Clear()
	e.renderEditorTextOnScreen()
	e.screen.Show()
}
