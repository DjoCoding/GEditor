package editor

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"unicode"

	"github.com/gdamore/tcell/v2"
)

const (
	PROMPT_SCREEN_LINE_BEGIN = 28
	PROMPT_SCREEN_COL_BEGIN  = 0

	LINE_CELL_ROW = 30
	LINE_CELL_COL = 30

	UPPER_CURSOR_BOUNDS  = 3
	BOTTOM_CURSOR_BOUNDS = 3

	UP = iota
	DOWN
	RIGHT
	LEFT

	NORMAL_MODE = iota
	EXITING_MODE
	SEARCHING_MODE
)

type EditorConfiguration struct {
	Filepath *string
}

type EditorSearchModeParams struct {
	text      string // this is used for the search and replace
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
		mode:            NORMAL_MODE,
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
	return editor.mode != EXITING_MODE
}

// set the editor to the quitting mode
func (editor *Editor) Quit() {
	editor.mode = EXITING_MODE
}

// insert char in the editor buffer
func (editor *Editor) insertChar(c rune) error {
	return editor.buffer.InsertChar(c, &editor.realCursor)
}

// remove a char from the editor buffer
func (editor *Editor) removeChar() error {
	return editor.buffer.RemoveChar(&editor.realCursor)
}

// insert the new line char '\n' into the editor buffer
func (editor *Editor) insertNewLine() error {
	return editor.buffer.InsertNewLine(&editor.realCursor)
}

// insert the tab char '\t' into the editor buffer
func (editor *Editor) insertTab() error {
	return editor.buffer.InsertTab(&editor.realCursor)
}

// move the (main) editor cursor up
func (editor *Editor) moveCursorUp() {
	realCursorLine := editor.realCursor.GetLine()

	if realCursorLine == 0 {
		editor.realCursor.SetCol(0)
		return
	}

	editor.realCursor.SetLine(realCursorLine - 1)

	prevLineCount := editor.buffer.lines[realCursorLine-1].Count()
	if prevLineCount < editor.realCursor.GetCol() {
		editor.realCursor.SetCol(prevLineCount)
	}
}

// move the (main) editor cursor down
func (editor *Editor) moveCursorDown() {
	realCursorLine := editor.realCursor.GetLine()

	if realCursorLine == editor.buffer.Count()-1 {
		editor.realCursor.SetCol(editor.buffer.LastLineCount())
		return
	}

	editor.realCursor.SetLine(realCursorLine + 1)

	nextLineCount := editor.buffer.lines[realCursorLine+1].Count()
	if nextLineCount < editor.realCursor.GetCol() {
		editor.realCursor.SetCol(nextLineCount)
	}
}

// move the (main) editor cursor left
func (editor *Editor) moveCursorLeft() {
	realCursorCol := editor.realCursor.GetCol()
	if realCursorCol > 0 {
		editor.realCursor.SetCol(realCursorCol - 1)
		return
	}

	realCursorLine := editor.realCursor.GetLine()
	if realCursorLine == 0 {
		return
	}

	editor.realCursor.SetLine(realCursorLine - 1)
	editor.realCursor.SetCol(editor.buffer.lines[editor.realCursor.GetLine()].Count())
}

// move the (main) editor cursor right
func (editor *Editor) moveCursorRight() {
	realCursorLine, realCursorCol := editor.realCursor.Get()

	if realCursorCol < editor.buffer.lines[realCursorLine].Count() {
		editor.realCursor.SetCol(realCursorCol + 1)
		return
	}

	if realCursorLine >= editor.buffer.Count()-1 {
		return
	}

	editor.realCursor.SetLine(realCursorLine + 1)
	editor.realCursor.SetCol(0)
}

// get the char at the location before the current cursor position
func (editor *Editor) getCharBeforeCursor() (c rune, ok bool) {
	line, col := editor.realCursor.Get()
	if col == 0 && line == 0 {
		return 0, false
	}

	for col == 0 {
		return '\n', true
	}

	return rune(editor.buffer.lines[line].content[col-1]), true
}

// get the char at the location after the current cursor position
func (editor *Editor) getCharAfterCursor() (c rune, ok bool) {
	line, col := editor.realCursor.Get()
	if col >= editor.buffer.lines[line].Count()-1 {
		if line >= editor.buffer.Count()-1 {
			return 0, false
		}

		return '\n', true
	}

	return rune(editor.buffer.lines[line].content[col+1]), true
}

// skip the token at the left position from the cursor
func (editor *Editor) skipLeftToken() {
	c, ok := editor.getCharBeforeCursor()
	if !ok {
		return
	}

	hasNewLine := false
	if unicode.IsSpace(c) {
		for unicode.IsSpace(c) {
			if c == '\n' {
				if hasNewLine {
					return
				}
				hasNewLine = true
			}

			editor.moveCursorLeft()
			c, ok = editor.getCharBeforeCursor()
			if !ok {
				return
			}
		}
	}

	if unicode.IsLetter(c) || c == '_' {
		for unicode.IsLetter(c) || c == '_' {
			editor.moveCursorLeft()
			c, ok = editor.getCharBeforeCursor()
			if !ok {
				return
			}
		}

		return
	}

	if unicode.IsNumber(c) {
		for unicode.IsNumber(c) {
			editor.moveCursorLeft()
			c, ok = editor.getCharBeforeCursor()
			if !ok {
				return
			}
		}

		return
	}

	editor.moveCursorLeft()
}

// skip the token at the right position from the cursor
func (editor *Editor) skipRightToken() {
	c, ok := editor.getCharAfterCursor()
	if !ok {
		return
	}

	hasNewLine := false
	if unicode.IsSpace(c) {
		for unicode.IsSpace(c) {
			if c == '\n' {
				if hasNewLine {
					return
				}
				hasNewLine = true
			}

			editor.moveCursorRight()
			c, ok = editor.getCharAfterCursor()
			if !ok {
				return
			}
		}
	}

	if unicode.IsLetter(c) || c == '_' {
		for unicode.IsLetter(c) || c == '_' {
			editor.moveCursorRight()
			c, ok = editor.getCharAfterCursor()
			if !ok {
				return
			}
		}

		return
	}

	if unicode.IsNumber(c) {
		for unicode.IsNumber(c) {
			editor.moveCursorRight()
			c, ok = editor.getCharAfterCursor()
			if !ok {
				return
			}
		}

		return
	}

	editor.moveCursorRight()
}

// handle the `Ctrl` + `Key` commands in the normal mode
func (editor *Editor) handleCtrlCommandsInNormalMode(key tcell.Key) error {
	switch key {
	case tcell.KeyLeft:
		editor.skipLeftToken()
	case tcell.KeyRight:
		editor.skipRightToken()
		// extra right moving (vscode mode)
		editor.moveCursorRight()
	case tcell.KeyCtrlS:
		err := editor.save()
		return err
	case tcell.KeyCtrlF:
		editor.mode = SEARCHING_MODE
	default:
		break
	}

	return nil
}

// handle the normal mode commands
func (editor *Editor) handleNormalModeEvent(ev tcell.Event) error {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		// handle the ctrl + `key` commands
		if ev.Modifiers()&tcell.ModCtrl != 0 {
			return editor.handleCtrlCommandsInNormalMode(ev.Key())
		}

		switch {
		case ev.Key() == tcell.KeyEscape:
			editor.quitAndSave()
		case ev.Key() == tcell.KeyBackspace2:
			editor.removeChar()
		case ev.Key() == tcell.KeyTab:
			editor.insertTab()
		case ev.Key() == tcell.KeyEnter:
			editor.insertNewLine()
		case ev.Key() == tcell.KeyUp:
			editor.moveCursorUp()
		case ev.Key() == tcell.KeyDown:
			editor.moveCursorDown()
		case ev.Key() == tcell.KeyLeft:
			editor.moveCursorLeft()
		case ev.Key() == tcell.KeyRight:
			editor.moveCursorRight()
		case ev.Key() == tcell.KeyRune:
			return editor.insertChar(ev.Rune())
		default:
			break
		}
	}

	return nil
}

// search for the text given in the search params (field in the editor) and set the cursor to its position
func (editor *Editor) searchAndUpdateCursor() {
	// reset the search pointer
	editor.searchParams.current = 0

	// update the search locations
	editor.updateSearchLocations(editor.searchParams.text)
	if len(editor.searchParams.locations) == 0 {
		return
	}

	// set the real cursor
	row, col := editor.searchParams.locations[editor.searchParams.current].Get()
	editor.realCursor = NewLocation(row, col+len(editor.searchParams.text))
}

// get the next position of the cursor from the current matching word (search function)
func (editor *Editor) updateSearchPointer() {
	locationsLen := len(editor.searchParams.locations)
	if locationsLen == 0 {
		return
	}

	// incrementing the pointer
	editor.searchParams.current++
	editor.searchParams.current %= locationsLen

	// updating the real cursor
	row, col := editor.searchParams.locations[editor.searchParams.current].Get()
	editor.realCursor = NewLocation(row, col+len(editor.searchParams.text))
}

// get back to the normal mode from the search mode
func (editor *Editor) switchToNormalFromSearchMode() {
	editor.searchParams.text = ""
	editor.searchParams.locations = nil
	editor.mode = NORMAL_MODE
}

// remove a char from the text in the search mode
func (editor *Editor) removeCharFromSearchModeText() {
	if len(editor.searchParams.text) == 0 {
		return
	}
	editor.searchParams.text = editor.searchParams.text[:len(editor.searchParams.text)-1]
}

// lookup a location in all the locations of the matching positions (after the search)
func (editor *Editor) lookupLocationInSearchLocations(loc Location) bool {
	for _, location := range editor.searchParams.locations {
		if location.Cmp(loc) {
			return true
		}
	}

	return false
}

// insert a char into the current searched text
func (editor *Editor) insertCharToSearchedText(c rune) {
	editor.searchParams.text += string(c)
}

// reset the current search test
func (editor *Editor) resetSearchedText() {
	editor.searchParams.text = ""
}

// handle search mode commands
func (editor *Editor) handleSearchModeEvent(ev tcell.Event) error {
	shouldMakeSearch := false

	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch {
		case ev.Key() == tcell.KeyEscape:
			editor.switchToNormalFromSearchMode()
		case ev.Key() == tcell.KeyBackspace2:
			editor.removeCharFromSearchModeText()
			shouldMakeSearch = true
		case ev.Key() == tcell.KeyEnter:
			editor.updateSearchPointer()
		case ev.Key() == tcell.KeyCtrlR:
			editor.resetSearchedText()
			shouldMakeSearch = true
		case ev.Key() == tcell.KeyRune:
			editor.insertCharToSearchedText(ev.Rune())
			shouldMakeSearch = true
		}
	}

	if shouldMakeSearch {
		editor.searchAndUpdateCursor()
	}

	return nil
}

// get an event from the event loop
func (editor *Editor) PollEvent() tcell.Event {
	return editor.screen.PollEvent()
}

// handle the event
func (editor *Editor) HandleEvent(ev tcell.Event) error {
	switch editor.mode {
	case NORMAL_MODE:
		return editor.handleNormalModeEvent(ev)
	case SEARCHING_MODE:
		return editor.handleSearchModeEvent(ev)
	}

	return nil
}

func (editor *Editor) renderLineInNormalMode(lineIndex int, row int) {
	line := editor.buffer.lines[lineIndex]

	for i, c := range line.GetContent() {
		editor.screen.SetContent(i, row, c, nil, tcell.StyleDefault)
	}
}

func (editor *Editor) renderLineInSearchMode(lineIndex int, row int) {
	style := tcell.StyleDefault.Bold(true).Underline(true).Background(tcell.ColorDarkCyan)
	line := editor.buffer.lines[lineIndex]

	count := 0

	for i, c := range line.GetContent() {
		currentLocation := NewLocation(lineIndex, i)
		found := editor.lookupLocationInSearchLocations(currentLocation)

		if found {
			count = len(editor.searchParams.text)
		}

		if count > 0 {
			editor.screen.SetContent(i, row, c, nil, style)
			count--
			continue
		}

		editor.screen.SetContent(i, row, c, nil, tcell.StyleDefault)
	}
}

func (editor *Editor) updateRenderingCursor() {
	for editor.realCursor.GetLine() < editor.renderingCursor.GetLine()+UPPER_CURSOR_BOUNDS {
		editor.renderingCursor.SetLine(editor.renderingCursor.GetLine() - 1)
		if editor.renderingCursor.GetLine() < 0 {
			editor.renderingCursor.SetLine(0)
			return
		}
	}

	_, h := editor.screen.Size()
	h -= BOTTOM_CURSOR_BOUNDS
	for editor.realCursor.GetLine() > editor.renderingCursor.GetLine()+h-BOTTOM_CURSOR_BOUNDS {
		editor.renderingCursor.SetLine(editor.renderingCursor.GetLine() + 1)
	}
}

func (editor *Editor) updateRelativeCursor() {
	editor.relativeCursor.Set(editor.realCursor.GetLine()-editor.renderingCursor.GetLine(), editor.realCursor.GetCol()-editor.renderingCursor.GetCol())
}

func (editor *Editor) getNumberLinesToRender() int {
	_, h := editor.screen.Size()
	h -= BOTTOM_CURSOR_BOUNDS

	return int(math.Min(float64(h), float64(editor.buffer.Count()-editor.renderingCursor.GetLine())))
}

// render the content of the editor buffer in the normal mode
func (editor *Editor) renderContentInNormalMode() {
	numberLinesToRender := editor.getNumberLinesToRender()
	for i := 0; i < numberLinesToRender; i++ {
		editor.renderLineInNormalMode(editor.renderingCursor.GetLine()+i, i)
	}
}

// render the content of the editor buffer in the search mode
func (editor *Editor) renderContentInSearchMode() {
	numberLinesToRender := editor.getNumberLinesToRender()
	for i := 0; i < numberLinesToRender; i++ {
		editor.renderLineInSearchMode(editor.renderingCursor.GetLine()+i, i)
	}
}

// render the content of the editor buffer
func (editor *Editor) renderContent() {
	editor.updateRenderingCursor()

	switch editor.mode {
	case NORMAL_MODE:
		editor.renderContentInNormalMode()
	case SEARCHING_MODE:
		editor.renderContentInSearchMode()
	}
}

// render the cursor of the editor (real Cursor)
func (editor *Editor) renderCursor() {
	editor.updateRelativeCursor()
	editor.screen.ShowCursor(editor.relativeCursor.GetCol(), editor.relativeCursor.GetLine())
}

// render any text to the editor screen (helper function)
func (editor *Editor) renderText(line, col int, text string) {
	for i, c := range text {
		editor.screen.SetContent(col+i, line, c, nil, tcell.StyleDefault)
	}
}

// render information (mode, cursor)
func (editor *Editor) renderInfo() {
	lineString := strconv.Itoa(editor.realCursor.GetLine())
	colString := strconv.Itoa(editor.realCursor.GetCol())

	editor.renderText(LINE_CELL_ROW, LINE_CELL_COL, lineString)
	editor.renderText(LINE_CELL_ROW, LINE_CELL_COL+len(lineString), ":")
	editor.renderText(LINE_CELL_ROW, LINE_CELL_COL+len(lineString)+1, colString)

	if editor.mode == SEARCHING_MODE {
		editor.renderText(PROMPT_SCREEN_LINE_BEGIN+1, PROMPT_SCREEN_COL_BEGIN, "find text: "+editor.searchParams.text)
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

// load a file using the EditorConfiguration fields (passed as args)
func (editor *Editor) loadFileFromConfiguration() error {
	fileInfo, err := os.Stat(*editor.config.Filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if fileInfo.IsDir() {
		return fmt.Errorf("can not open directories in this text editor")
	}

	fileContent, err := os.ReadFile(*editor.config.Filepath)
	if err != nil {
		return err
	}

	for _, c := range fileContent {
		switch c {
		case '\n':
			err = editor.insertNewLine()
		case '\t':
			err = editor.insertTab()
		default:
			err = editor.insertChar(rune(c))
		}

		if err != nil {
			return err
		}
	}

	return nil
}

// load file to the editor buffer
// main function
func (editor *Editor) Load() error {
	if editor.config.Filepath == nil {
		return nil
	}

	return editor.loadFileFromConfiguration()
}

// save the content of the editor buffer to a file
// main function
func (editor *Editor) saveContent(f *os.File) error {
	for _, line := range editor.buffer.lines {
		_, err := f.Write([]byte(line.content))
		if err != nil {
			return err
		}
		_, err = f.Write([]byte("\n"))
		if err != nil {
			return err
		}
	}

	return nil
}

// not implemented yet
func (editor *Editor) saveFromConfiguration() error {
	return nil
}

// save into a hardcoded filepath
func (editor *Editor) save() error {
	if editor.config.Filepath != nil {
		return editor.saveFromConfiguration()
	}

	filepath := "./test"
	f, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	return editor.saveContent(f)
}

// quit the editor and save into a hardcoded filepath
func (editor *Editor) quitAndSave() error {
	editor.Quit()
	return editor.save()
}

// search a text in the editor buffer and set all the locations where found
func (editor *Editor) updateSearchLocations(text string) {
	editor.searchParams.locations = editor.buffer.Search(editor.realCursor, text)
}
