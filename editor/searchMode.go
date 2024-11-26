package editor

import "github.com/gdamore/tcell/v2"

func (editor *Editor) setSearchMode() {
	editor.searchParams = EditorSearchModeParams{}
	editor.mode = SEARCH_MODE
	editor.enableInputBuffer()
	editor.setInputCurrentBuffer(INPUT_TEXT)
	editor.setInputBufferInputRequestString("find text: ")
}

func (editor *Editor) setSearchSubMode(whichMode int) {
	editor.searchParams.whichMode = whichMode
}

// get back to the normal mode from the search mode
func (editor *Editor) switchToNormalFromSearchMode() {
	editor.searchParams = EditorSearchModeParams{}
	editor.input = EditorInternalInput{}
	editor.mode = INSERT_MODE
}

// search for the text given in the search params (field in the editor) and set the cursor to its position
func (editor *Editor) searchAndSetCursor() {
	// reset the search pointer
	editor.searchParams.current = 0

	// update the search locations
	editor.updateSearchLocations(editor.input.buffers[INPUT_TEXT])
	if len(editor.searchParams.locations) == 0 {
		return
	}

	// set the real cursor
	row, col := editor.searchParams.locations[editor.searchParams.current].Get()
	editor.realCursor = NewLocation(row, col+len(editor.input.buffers[INPUT_TEXT]))
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
	editor.realCursor = NewLocation(row, col+len(editor.input.buffers[INPUT_TEXT]))
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

// search a text in the editor buffer and set all the locations where found
func (editor *Editor) updateSearchLocations(text string) {
	editor.searchParams.locations = editor.buffer.Search(editor.realCursor, text)
}

func (e *Editor) replaceOnCursor() {
	if len(e.searchParams.locations) == 0 {
		e.switchToNormalFromSearchMode()
		return
	}

	newText := e.input.buffers[NEW_TEXT]
	oldText := e.input.buffers[INPUT_TEXT]
	e.realCursor.SetCol(e.realCursor.GetCol() - len(oldText))

	e.buffer.findAndReplace(newText, oldText, &e.realCursor)
	e.searchParams.hasReplaced = true

	currentLocation := e.realCursor
	e.searchAndSetCursor()

	e.realCursor = currentLocation
}

func (editor *Editor) handleEscapeKeyInSearchMode() {
	mode := editor.searchParams.whichMode

	if mode == SEARCH {
		editor.switchToNormalFromSearchMode()
		return
	}

	if editor.getInputCurrentBuffer() == INPUT_TEXT {
		editor.switchToNormalFromSearchMode()
		return
	}

	editor.setInputCurrentBuffer(INPUT_TEXT)
}

func (editor *Editor) handleEnterKeyInSearchMode() {
	mode := editor.searchParams.whichMode

	if mode == SEARCH {
		editor.updateSearchPointer()
		return
	}

	if editor.getInputCurrentBuffer() == INPUT_TEXT {
		editor.setInputCurrentBuffer(NEW_TEXT)
		editor.setInputBufferInputRequestString("replace with: ")
		return
	}

	if !editor.searchParams.hasReplaced {
		editor.replaceOnCursor()
		return
	}

	editor.searchAndSetCursor()
	editor.searchParams.hasReplaced = false
}

// handle search mode commands
func (editor *Editor) handleSearchModeEvent(ev tcell.Event) error {
	shouldMakeSearch := false

	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch {
		case ev.Key() == tcell.KeyEscape:
			editor.handleEscapeKeyInSearchMode()
		case ev.Key() == tcell.KeyBackspace2:
			editor.removeCharFromInputBuffer()
			shouldMakeSearch = true
		case ev.Key() == tcell.KeyEnter:
			editor.handleEnterKeyInSearchMode()
		case ev.Key() == tcell.KeyRune:
			editor.insertCharToInputBuffer(ev.Rune())
			shouldMakeSearch = true
		}
	}

	if shouldMakeSearch {
		editor.searchAndSetCursor()
	}

	return nil
}
