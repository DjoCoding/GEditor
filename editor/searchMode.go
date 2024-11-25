package editor

import "github.com/gdamore/tcell/v2"

func (editor *Editor) setSearchMode() {
	editor.searchParams = EditorSearchModeParams{}
	editor.mode = SEARCH_MODE
}

func (editor *Editor) setSearchSubMode(whichMode int) {
	editor.searchParams.whichMode = whichMode
}

// get back to the normal mode from the search mode
func (editor *Editor) switchToNormalFromSearchMode() {
	editor.searchParams = EditorSearchModeParams{}
	editor.mode = INSERT_MODE
}

func (editor *Editor) setSearchModeCurrentBuffer(whichBuffer int) {
	editor.searchParams.currentBuffer = whichBuffer
}

// search for the text given in the search params (field in the editor) and set the cursor to its position
func (editor *Editor) searchAndSetCursor() {
	// reset the search pointer
	editor.searchParams.current = 0

	// update the search locations
	editor.updateSearchLocations(editor.searchParams.buffers[ON_WHICHTEXT])
	if len(editor.searchParams.locations) == 0 {
		return
	}

	// set the real cursor
	row, col := editor.searchParams.locations[editor.searchParams.current].Get()
	editor.realCursor = NewLocation(row, col+len(editor.searchParams.buffers[ON_WHICHTEXT]))
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
	editor.realCursor = NewLocation(row, col+len(editor.searchParams.buffers[ON_WHICHTEXT]))
}

// remove a char from the text in the search mode
func (editor *Editor) removeCharFromSearchModeText() {
	if len(editor.searchParams.buffers[ON_WHICHTEXT]) == 0 {
		return
	}
	editor.searchParams.buffers[ON_WHICHTEXT] = editor.searchParams.buffers[ON_WHICHTEXT][:len(editor.searchParams.buffers[ON_WHICHTEXT])-1]
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
	editor.searchParams.buffers[editor.searchParams.currentBuffer] += string(c)
}

// search a text in the editor buffer and set all the locations where found
func (editor *Editor) updateSearchLocations(text string) {
	editor.searchParams.locations = editor.buffer.Search(editor.realCursor, text)
}

func (e *Editor) replaceOnCursor() {
	if len(e.searchParams.locations) == 0 {
		e.switchToNormalFromSearchMode()
		return
	}

	newText := e.searchParams.buffers[ON_NEWTEXT]
	oldText := e.searchParams.buffers[ON_WHICHTEXT]
	e.realCursor.SetCol(e.realCursor.GetCol() - len(oldText))

	e.buffer.findAndReplace(newText, oldText, &e.realCursor)
	e.searchParams.hasReplaced = true

	currentLocation := e.realCursor
	e.searchAndSetCursor()

	e.realCursor = currentLocation
}

// handle search mode commands
func (editor *Editor) handleSearchModeEvent(ev tcell.Event) error {
	shouldMakeSearch := false

	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch {
		case ev.Key() == tcell.KeyEscape:
			switch editor.searchParams.whichMode {
			case REPLACE:
				switch editor.searchParams.currentBuffer {
				case ON_NEWTEXT:
					editor.searchParams.currentBuffer = ON_WHICHTEXT
				default:
					editor.switchToNormalFromSearchMode()
				}

			default:
				editor.switchToNormalFromSearchMode()
			}

		case ev.Key() == tcell.KeyBackspace2:
			editor.removeCharFromSearchModeText()
			shouldMakeSearch = true

		case ev.Key() == tcell.KeyEnter:
			switch editor.searchParams.whichMode {
			case SEARCH:
				editor.updateSearchPointer()
			case REPLACE:
				if editor.searchParams.currentBuffer == ON_NEWTEXT {
					switch editor.searchParams.hasReplaced {
					case true:
						editor.searchAndSetCursor()
						editor.searchParams.hasReplaced = false
					default:
						editor.replaceOnCursor()
					}

				} else {
					editor.setSearchModeCurrentBuffer(ON_NEWTEXT)
				}
			}

		case ev.Key() == tcell.KeyRune:
			editor.insertCharToSearchedText(ev.Rune())
			shouldMakeSearch = true
		}
	}

	if shouldMakeSearch {
		editor.searchAndSetCursor()
	}

	return nil
}
