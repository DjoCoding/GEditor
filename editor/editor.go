package editor

import (
	"fmt"
	"os"
	"strconv"

	"github.com/gdamore/tcell/v2"
)

const (
	LINE_CELL_ROW = 20
	LINE_CELL_COL = 30

	EDITOR_TAB_SIZE = 4

	UP = iota
	DOWN
	RIGHT
	LEFT
)

type EditorConfiguration struct {
	Filepath *string
}

type Editor struct {
	screen tcell.Screen
	buffer Buffer
	cursor Cursor
	quit   bool
	config EditorConfiguration
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
		screen: screen,
		buffer: NewBuffer(),
		cursor: NewCursor(),
		quit:   false,
		config: editorConfig,
	}, nil
}

func (editor *Editor) Close() {
	editor.screen.Fini()
}

func (editor *Editor) ShouldNotQuit() bool {
	return !editor.quit
}

func (editor *Editor) Quit() {
	editor.quit = true
}

func (editor *Editor) insertChar(c rune) error {
	return editor.buffer.InsertChar(c, &editor.cursor)
}

func (editor *Editor) removeChar() error {
	return editor.buffer.RemoveChar(&editor.cursor)
}

func (editor *Editor) insertNewLine() error {
	return editor.buffer.InsertNewLine(&editor.cursor)
}

func (editor *Editor) insertTab() error {
	return editor.buffer.InsertTab(&editor.cursor)
}

func (editor *Editor) moveCursorUp() {
	cursorLine := editor.cursor.GetLine()

	if cursorLine == 0 {
		editor.cursor.SetCol(0)
		return
	}

	editor.cursor.SetLine(cursorLine - 1)

	prevLineCount := editor.buffer.lines[cursorLine-1].Count()
	if prevLineCount < editor.cursor.GetCol() {
		editor.cursor.SetCol(prevLineCount)
	}
}

func (editor *Editor) moveCursorDown() {
	cursorLine := editor.cursor.GetLine()

	if cursorLine == editor.buffer.Count()-1 {
		editor.cursor.SetCol(editor.buffer.LastLineCount())
		return
	}

	editor.cursor.SetLine(cursorLine + 1)

	nextLineCount := editor.buffer.lines[cursorLine+1].Count()
	if nextLineCount < editor.cursor.GetCol() {
		editor.cursor.SetCol(nextLineCount)
	}
}

func (editor *Editor) moveCursorLeft() {
	cursorCol := editor.cursor.GetCol()
	if cursorCol > 0 {
		editor.cursor.SetCol(cursorCol - 1)
		return
	}

	cursorLine := editor.cursor.GetLine()
	if cursorLine == 0 {
		return
	}

	editor.cursor.SetLine(cursorLine - 1)
	editor.cursor.SetCol(editor.buffer.lines[editor.cursor.GetLine()].Count())
}

func (editor *Editor) moveCursorRight() {
	cursorLine, cursorCol := editor.cursor.Get()

	if cursorCol < editor.buffer.lines[cursorLine].Count() {
		editor.cursor.SetCol(cursorCol + 1)
		return
	}

	if cursorLine >= editor.buffer.Count()-1 {
		return
	}

	editor.cursor.SetLine(cursorLine + 1)
	editor.cursor.SetCol(0)
}

func (editor *Editor) PollEvent() tcell.Event {
	return editor.screen.PollEvent()
}

func (editor *Editor) HandleEvent(ev tcell.Event) error {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch {
		case ev.Key() == tcell.KeyEscape:
			editor.Quit()
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
		default:
			return editor.insertChar(ev.Rune())
		}
	}

	return nil
}

func (editor *Editor) renderLine(lineIndex int, row int) {
	line := editor.buffer.lines[lineIndex]

	for i, c := range line.GetContent() {
		editor.screen.SetContent(i, row, c, nil, tcell.StyleDefault)
	}
}

func (editor *Editor) renderContent() {
	for i := 0; i < len(editor.buffer.lines); i++ {
		editor.renderLine(i, i)
	}
}

func (editor *Editor) renderCursor() {
	editor.screen.ShowCursor(editor.cursor.GetCol(), editor.cursor.GetLine())
}

func (editor *Editor) renderInfo() {
	lineString := strconv.Itoa(editor.cursor.GetLine())
	colString := strconv.Itoa(editor.cursor.GetCol())

	for i, c := range lineString {
		editor.screen.SetContent(LINE_CELL_COL+i, LINE_CELL_ROW, c, nil, tcell.StyleDefault)
	}

	editor.screen.SetContent(LINE_CELL_COL+len(lineString), LINE_CELL_ROW, ':', nil, tcell.StyleDefault)

	for i, c := range colString {
		editor.screen.SetContent(LINE_CELL_COL+len(lineString)+i+1, LINE_CELL_ROW, c, nil, tcell.StyleDefault)
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
