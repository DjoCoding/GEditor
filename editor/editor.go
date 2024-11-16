package editor

import (
	"fmt"
	"math"
	"os"
	"strconv"

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

	RUNNING_MODE = iota
	EXITING_MODE
	SEARCHING_MODE
)

type EditorConfiguration struct {
	Filepath *string
}

type Editor struct {
	screen          tcell.Screen
	buffer          Buffer
	realCursor      Location
	relativeCursor  Location
	renderingCursor Location
	config          EditorConfiguration
	mode            int
	text            string // this is used for the search and replace
}

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
		realCursor:      NewLocation(),
		relativeCursor:  NewLocation(),
		renderingCursor: NewLocation(),
		mode:            RUNNING_MODE,
		config:          editorConfig,
	}, nil
}

func (editor *Editor) Close() {
	editor.screen.Fini()
}

func (editor *Editor) ShouldNotQuit() bool {
	return editor.mode != EXITING_MODE
}

func (editor *Editor) Quit() {
	editor.mode = EXITING_MODE
}

func (editor *Editor) insertChar(c rune) error {
	return editor.buffer.InsertChar(c, &editor.realCursor)
}

func (editor *Editor) removeChar() error {
	return editor.buffer.RemoveChar(&editor.realCursor)
}

func (editor *Editor) insertNewLine() error {
	return editor.buffer.InsertNewLine(&editor.realCursor)
}

func (editor *Editor) insertTab() error {
	return editor.buffer.InsertTab(&editor.realCursor)
}

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

func (editor *Editor) PollEvent() tcell.Event {
	return editor.screen.PollEvent()
}

func (editor *Editor) handleNormalModeEvent(ev tcell.Event) error {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch {
		case ev.Key() == tcell.KeyEscape:
			editor.QuitAndSave()
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
		case ev.Key() == tcell.KeyCtrlS:
			err := editor.Save()
			return err
		case ev.Key() == tcell.KeyCtrlF:
			editor.mode = SEARCHING_MODE
		default:
			return editor.insertChar(ev.Rune())
		}
	}

	return nil
}

func (editor *Editor) SearchAndUpdateCursor() {
	newCursor, found := editor.Search(editor.text)
	if !found {
		return
	}
	editor.realCursor = newCursor
}

func (editor *Editor) handleSearchModeEvent(ev tcell.Event) error {
	shouldMakeSearch := false

	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch {
		case ev.Key() == tcell.KeyEscape:
			editor.text = ""
			editor.mode = RUNNING_MODE
		case ev.Key() == tcell.KeyBackspace2:
			if len(editor.text) != 0 {
				editor.text = editor.text[:len(editor.text)-1]
				shouldMakeSearch = true
			}
		case ev.Key() == tcell.KeyEnter:
			shouldMakeSearch = true
		default:
			editor.text += string(ev.Rune())
			shouldMakeSearch = true
		}
	}

	if shouldMakeSearch {
		editor.SearchAndUpdateCursor()
	}

	return nil
}

func (editor *Editor) HandleEvent(ev tcell.Event) error {
	switch editor.mode {
	case RUNNING_MODE:
		return editor.handleNormalModeEvent(ev)
	case SEARCHING_MODE:
		return editor.handleSearchModeEvent(ev)
	}

	return nil
}

func (editor *Editor) renderLine(lineIndex int, row int) {
	line := editor.buffer.lines[lineIndex]

	for i, c := range line.GetContent() {
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

func (editor *Editor) renderContent() {
	editor.updateRenderingCursor()

	_, h := editor.screen.Size()
	h -= BOTTOM_CURSOR_BOUNDS

	numberLinesToRender := int(math.Min(float64(h), float64(editor.buffer.Count()-editor.renderingCursor.GetLine())))

	for i := 0; i < numberLinesToRender; i++ {
		editor.renderLine(editor.renderingCursor.GetLine()+i, i)
	}
}

func (editor *Editor) renderCursor() {
	editor.updateRelativeCursor()
	editor.screen.ShowCursor(editor.relativeCursor.GetCol(), editor.relativeCursor.GetLine())
}

func (editor *Editor) renderText(line, col int, text string) {
	for i, c := range text {
		editor.screen.SetContent(col+i, line, c, nil, tcell.StyleDefault)
	}
}

func (editor *Editor) renderInfo() {
	lineString := strconv.Itoa(editor.realCursor.GetLine())
	colString := strconv.Itoa(editor.realCursor.GetCol())

	editor.renderText(LINE_CELL_ROW, LINE_CELL_COL, lineString)
	editor.renderText(LINE_CELL_ROW, LINE_CELL_COL+len(lineString), ":")
	editor.renderText(LINE_CELL_ROW, LINE_CELL_COL+len(lineString)+1, colString)

	if editor.mode == SEARCHING_MODE {
		editor.renderText(PROMPT_SCREEN_LINE_BEGIN+1, PROMPT_SCREEN_COL_BEGIN, "text: "+editor.text)
	}
}

func (editor *Editor) Render() {
	editor.screen.Clear()
	editor.renderContent()
	editor.renderInfo()
	editor.renderCursor()
	editor.screen.Show()
}

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

func (editor *Editor) Load() error {
	if editor.config.Filepath == nil {
		return nil
	}

	return editor.loadFileFromConfiguration()
}

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

func (editor *Editor) saveFromConfiguration() error {
	return nil
}

func (editor *Editor) Save() error {
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

func (editor *Editor) QuitAndSave() error {
	editor.Quit()
	return editor.Save()
}

func (editor *Editor) Search(text string) (location Location, found bool) {
	return editor.buffer.Search(editor.realCursor, text)
}
