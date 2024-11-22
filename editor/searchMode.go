package editor

import "github.com/gdamore/tcell"

// search for the text given in the search params (field in the editor) and set the cursor to its position
func (editor *Editor) searchAndUpdateCursor() {
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

// get back to the normal mode from the search mode
func (editor *Editor) switchToInsertFromSearchMode() {
	editor.searchParams = EditorSearchModeParams{}
	editor.setMode(INSERT_MODE)
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
	editor.searchParams.buffers[ON_WHICHTEXT] += string(c)
}

// search a text in the editor buffer and set all the locations where found
func (editor *Editor) updateSearchLocations(text string) {
	editor.searchParams.locations = editor.buffer.Search(editor.realCursor, text)
}

// handle search mode commands
func (editor *Editor) handleSearchModeEvent(ev tcell.Event) error {
	shouldMakeSearch := false

	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch {
		case ev.Key() == tcell.KeyEscape:
			editor.switchToInsertFromSearchMode()
		case ev.Key() == tcell.KeyBackspace2:
			editor.removeCharFromSearchModeText()
			shouldMakeSearch = true
		case ev.Key() == tcell.KeyEnter:
			editor.updateSearchPointer()
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
