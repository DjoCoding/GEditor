package editor

import "github.com/gdamore/tcell/v2"

// get an event from the event loop
func (editor *Editor) PollEvent() tcell.Event {
	return editor.screen.PollEvent()
}

// handle the event
func (editor *Editor) HandleEvent(ev tcell.Event) error {
	switch editor.mode {
	case INSERT_MODE:
		return editor.handleInsertModeEvent(ev)
	case SEARCH_MODE:
		return editor.handleSearchModeEvent(ev)
	case SELECTION_MODE:
		return editor.handleSelectionModeEvent(ev)
	}

	return nil
}
