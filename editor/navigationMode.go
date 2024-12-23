package editor

import (
	"os"

	"github.com/gdamore/tcell/v2"
)

func (e *Editor) openDir(dirName string) error {
	files, err := os.ReadDir(dirName)
	if err != nil {
		return err
	}

	e.navParams.files = files
	return nil
}

func (e *Editor) setNavigationMode() {
	e.mode = NAVIGATION_MODE
	e.openDir(e.config.OpenedFile)
	e.navParams.currentFileIndex = 0
}

func (e *Editor) setNavigationModeFromInsertMode() {
	e.mode = NAVIGATION_MODE

	e.navParams.currentFileIndex = 0
	e.navParams.files = nil

	if e.config.OpenedFile == "" {
		e.enableInputBuffer()
		e.setInputCurrentBuffer(INPUT_TEXT)
		e.setInputBufferInputRequestString("filepath: ")
		return
	}

	e.openDir(e.config.OpenedFile)
}

func (e *Editor) updateFileIndexCursorDown() {
	if len(e.navParams.files) == 0 {
		return
	}

	e.navParams.currentFileIndex++
	e.navParams.currentFileIndex %= len(e.navParams.files)
}

func (e *Editor) updateFileIndexCursorUp() {
	if len(e.navParams.files) == 0 {
		return
	}

	count := len(e.navParams.files)

	if e.navParams.currentFileIndex == 0 {
		e.navParams.currentFileIndex = count - 1
		return
	}

	e.navParams.currentFileIndex--
}

func (e *Editor) handleEnterKeyInNavigationMode() error {
	filepath := e.navParams.files[e.navParams.currentFileIndex].Name()
	e.config.OpenedFile += "/" + filepath
	return e.loadFileFromConfiguration()
}

func (e *Editor) handleRuneKeyInNavigationMode(c rune) {
	for i, file := range e.navParams.files {
		if file.Name()[0] == byte(c) {
			e.navParams.currentFileIndex = i
			return
		}
	}
}

func (e *Editor) handleNavigationModeEvent(ev tcell.Event) error {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyEscape:
			e.Quit()
		case tcell.KeyDown:
			e.updateFileIndexCursorDown()
		case tcell.KeyUp:
			e.updateFileIndexCursorUp()
		case tcell.KeyEnter:
			e.handleEnterKeyInNavigationMode()
		case tcell.KeyRune:
			e.handleRuneKeyInNavigationMode(ev.Rune())
			return nil
		}
	}
	return nil
}
