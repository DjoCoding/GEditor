package editor

import (
	"os"
	"unicode"

	"github.com/gdamore/tcell/v2"
)

// insert char in the editor buffer
func (editor *Editor) insertChar(c rune) error {
	return editor.buffer.insertChar(c, &editor.realCursor)
}

// remove a char from the editor buffer
func (editor *Editor) removeChar() error {
	return editor.buffer.removeChar(&editor.realCursor)
}

// insert the new line char '\n' into the editor buffer
func (editor *Editor) insertNewLine() error {
	return editor.buffer.insertNewLine(&editor.realCursor)
}

// insert the tab char '\t' into the editor buffer
func (editor *Editor) insertTab() error {
	return editor.buffer.insertTab(&editor.realCursor)
}

// move the (main) editor cursor up
func (editor *Editor) moveCursorUp() {
	realCursorLine := editor.realCursor.getLine()

	if realCursorLine == 0 {
		editor.realCursor.setCol(0)
		return
	}

	editor.realCursor.setLine(realCursorLine - 1)

	prevLineCount := editor.buffer.lines[realCursorLine-1].count()
	if prevLineCount < editor.realCursor.getCol() {
		editor.realCursor.setCol(prevLineCount)
	}
}

// move the (main) editor cursor down
func (editor *Editor) moveCursorDown() {
	realCursorLine := editor.realCursor.getLine()

	if realCursorLine == editor.buffer.count()-1 {
		editor.realCursor.setCol(editor.buffer.lastLineCount())
		return
	}

	editor.realCursor.setLine(realCursorLine + 1)

	nextLineCount := editor.buffer.lines[realCursorLine+1].count()
	if nextLineCount < editor.realCursor.getCol() {
		editor.realCursor.setCol(nextLineCount)
	}
}

// move the (main) editor cursor left
func (editor *Editor) moveCursorLeft() {
	realCursorCol := editor.realCursor.getCol()
	if realCursorCol > 0 {
		editor.realCursor.setCol(realCursorCol - 1)
		return
	}

	realCursorLine := editor.realCursor.getLine()
	if realCursorLine == 0 {
		return
	}

	editor.realCursor.setLine(realCursorLine - 1)
	editor.realCursor.setCol(editor.buffer.lines[editor.realCursor.getLine()].count())
}

// move the (main) editor cursor right
func (editor *Editor) moveCursorRight() {
	realCursorLine, realCursorCol := editor.realCursor.get()

	if realCursorCol < editor.buffer.lines[realCursorLine].count() {
		editor.realCursor.setCol(realCursorCol + 1)
		return
	}

	if realCursorLine >= editor.buffer.count()-1 {
		return
	}

	editor.realCursor.setLine(realCursorLine + 1)
	editor.realCursor.setCol(0)
}

// get the char at the location before the current cursor position
func (editor *Editor) getCharBeforeCursor() (c rune, ok bool) {
	line, col := editor.realCursor.get()
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
	line, col := editor.realCursor.get()
	if col >= editor.buffer.lines[line].count()-1 {
		if line >= editor.buffer.count()-1 {
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

func writeFile(text string) {
	os.WriteFile("djocoding", []byte(text), 0644)
}

func (editor *Editor) handleFileSavingInInsertMode() error {
	if editor.config.CurrentFile != "" {
		return editor.save()
	}

	editor.enableInputBuffer()
	editor.setInputCurrentBuffer(INPUT_TEXT)
	editor.setInputBufferInputRequestString("filepath: ")
	return nil
}

// handle the `Ctrl` + `Key` commands in the normal mode
func (editor *Editor) handleCtrlCommandsInInsertMode(evKey tcell.EventKey) error {
	switch evKey.Key() {
	case tcell.KeyLeft:
		editor.skipLeftToken()
	case tcell.KeyRight:
		editor.skipRightToken()
		// extra right moving (vscode mode)
		editor.moveCursorRight()
	case tcell.KeyCtrlS:
		return editor.handleFileSavingInInsertMode()
	case tcell.KeyCtrlF:
		editor.setSearchMode()
		editor.setSearchSubMode(SEARCH)
	case tcell.KeyCtrlR:
		editor.setSearchMode()
		editor.setSearchSubMode(REPLACE)
	case tcell.KeyCtrlP:
		editor.setNavigationModeFromInsertMode()
	default:
		break
	}

	return nil
}

func (editor *Editor) handleEnterKeyInInsertMode() error {
	if editor.inputBufferIsEnabled() {
		if editor.input.buffers[editor.getInputCurrentBuffer()] == "" {
			editor.disableInputBuffer()
			return nil
		}

		editor.config.CurrentFile = editor.input.buffers[editor.getInputCurrentBuffer()]
		err := editor.save()
		if err != nil {
			return err
		}

		editor.resetInput()
		return nil
	}

	return editor.insertNewLine()
}

func (editor *Editor) handleBackSpaceKeyInInsertMode() error {
	if editor.inputBufferIsEnabled() {
		editor.removeCharFromInputBuffer()
		return nil
	}

	return editor.removeChar()
}

func (editor *Editor) handleRuneKeyInInsertMode(c rune) error {
	if editor.inputBufferIsEnabled() {
		editor.insertCharToInputBuffer(c)
		return nil
	}
	return editor.insertChar(c)
}

// handle the normal mode commands
func (editor *Editor) handleInsertModeEvent(ev tcell.Event) error {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		if ev.Modifiers()&tcell.ModShift != 0 {
			editor.setSelectionMode()
			return editor.HandleEvent(ev)
		}

		// handle the ctrl + `evKey.Key()` commands
		if ev.Modifiers()&tcell.ModCtrl != 0 {
			return editor.handleCtrlCommandsInInsertMode(*ev)
		}

		switch {
		case ev.Key() == tcell.KeyEscape:
			return editor.saveAndquit()
		case ev.Key() == tcell.KeyBackspace2:
			return editor.handleBackSpaceKeyInInsertMode()
		case ev.Key() == tcell.KeyTab:
			return editor.insertTab()
		case ev.Key() == tcell.KeyEnter:
			return editor.handleEnterKeyInInsertMode()
		case ev.Key() == tcell.KeyUp:
			editor.moveCursorUp()
		case ev.Key() == tcell.KeyDown:
			editor.moveCursorDown()
		case ev.Key() == tcell.KeyLeft:
			editor.moveCursorLeft()
		case ev.Key() == tcell.KeyRight:
			editor.moveCursorRight()
		case ev.Key() == tcell.KeyRune:
			return editor.handleRuneKeyInInsertMode(ev.Rune())
		default:
			break
		}
	}

	return nil
}
