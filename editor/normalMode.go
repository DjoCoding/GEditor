package editor

import (
	"unicode"

	"github.com/gdamore/tcell/v2"
)

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
func (editor *Editor) handleCtrlCommandsInNormalMode(evKey tcell.EventKey) error {
	switch evKey.Key() {
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
		editor.setSearchMode()
		editor.setSearchSubMode(SEARCH)
	case tcell.KeyCtrlR:
		editor.setSearchMode()
		editor.setSearchSubMode(REPLACE)
	default:
		break
	}

	return nil
}

// handle the normal mode commands
func (editor *Editor) handleNormalModeEvent(ev tcell.Event) error {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		if ev.Modifiers()&tcell.ModShift != 0 {
			editor.setSelectionMode()
			return editor.HandleEvent(ev)
		}
		
		// handle the ctrl + `evKey.Key()` commands
		if ev.Modifiers()&tcell.ModCtrl != 0 {
			return editor.handleCtrlCommandsInNormalMode(*ev)
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
