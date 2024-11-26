package editor

import (
	"github.com/gdamore/tcell/v2"
)

func (e *Editor) setSelectionMode() {
	e.mode = SELECTION_MODE
	e.selParams.startLocation = e.realCursor
	e.selParams.endLocation = e.realCursor
}

func sortLocations(a, b Location) (Location, Location) {
	if a.line < b.line {
		return a, b
	}

	if a.line == b.line {
		if a.col < b.col {
			return a, b
		}

		return b, a
	}

	return b, a
}

func (e *Editor) checkLocationInSelectionModeBounds(loc Location) bool {
	start, end := sortLocations(e.selParams.startLocation, e.selParams.endLocation)
	if loc.line < start.line {
		return false
	}

	if loc.line > end.line {
		return false
	}

	if loc.line > start.line && loc.line < end.line {
		return true
	}

	if start.line == end.line {
		if loc.col < end.col && loc.col >= start.col {
			return true
		}

		return false
	}

	if loc.line == start.line {
		return loc.col >= start.col
	}

	return loc.col < end.col
}

func (e *Editor) countDistanceBetweenSelectionModeBounds() int {
	start, end := sortLocations(e.selParams.startLocation, e.selParams.endLocation)

	sline, scol := start.Get()
	eline, ecol := end.Get()

	if sline == eline {
		return ecol - scol
	}

	count := 0

	for i := sline + 1; i < eline-1; i++ {
		count += e.buffer.lines[i].Count()
	}

	count += e.buffer.lines[sline].Count() - scol
	count += ecol

	return count
}

func (e *Editor) moveCursorLeftInSelectionMode() {
	e.moveCursorLeft()
	e.selParams.endLocation = e.realCursor
}

func (e *Editor) moveCursorRightInSelectionMode() {
	e.moveCursorRight()
	e.selParams.endLocation = e.realCursor
}

func (e *Editor) removeContentInSelectionMode() error {
	_, end := sortLocations(e.selParams.startLocation, e.selParams.endLocation)
	err := e.buffer.RemoveString(e.countDistanceBetweenSelectionModeBounds(), &end)
	if err != nil {
		return err
	}

	e.realCursor = end
	return nil
}

func (e *Editor) switchToInsertFromSelectionMode() {
	e.selParams = EditorSelectionModeParams{}
	e.mode = INSERT_MODE
}

func (e *Editor) skipLeftTokenInSelectionMode() {
	e.skipLeftToken()
	e.selParams.endLocation = e.realCursor
}

func (e *Editor) skipRightTokenInSelectionMode() {
	e.skipRightToken()
	e.selParams.endLocation = e.realCursor
}

func (e *Editor) handleSelectionModeEvent(ev tcell.Event) error {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		if ev.Modifiers()&tcell.ModShift == 0 {
			switch ev.Key() {
			case tcell.KeyRune:
				err := e.removeContentInSelectionMode()
				if err != nil {
					return err
				}
				e.insertChar(ev.Rune())
			case tcell.KeyBackspace2:
				err := e.removeContentInSelectionMode()
				if err != nil {
					return err
				}
			}

			e.switchToInsertFromSelectionMode()
			return nil
		}

		if ev.Modifiers()&tcell.ModCtrl != 0 {
			switch ev.Key() {
			case tcell.KeyLeft:
				e.skipLeftTokenInSelectionMode()
			case tcell.KeyRight:
				e.skipRightTokenInSelectionMode()
			default:
				return nil
			}
		}

		switch ev.Key() {
		case tcell.KeyLeft:
			e.moveCursorLeftInSelectionMode()
		case tcell.KeyRight:
			e.moveCursorRightInSelectionMode()
		case tcell.KeyBackspace2:
			err := e.removeContentInSelectionMode()
			if err != nil {
				return err
			}
			e.switchToInsertFromSelectionMode()
		case tcell.KeyEscape:
			e.switchToInsertFromSelectionMode()
		case tcell.KeyRune:
			err := e.removeContentInSelectionMode()
			if err != nil {
				return err
			}
			e.switchToInsertFromSelectionMode()
			return e.insertChar(ev.Rune())
		}
	}

	return nil
}
