package editor

import "fmt"

type Line struct {
	content string
}

func newLine(content string) Line {
	return Line{
		content: content,
	}
}

func (line *Line) getContent() string {
	return line.content
}

func (line *Line) isValidLocation(cursor int) bool {
	return cursor <= line.count()
}

func (line *Line) insertString(s string, cursor *Location) error {
	if !line.isValidLocation(cursor.getCol()) {
		return fmt.Errorf("[LINE ERROR] invalid cursor position, failed to append string %s", s)
	}

	line.content = line.content[0:cursor.getCol()] + s + line.content[cursor.getCol():]
	cursor.setCol(cursor.getCol() + len(s))
	return nil
}

func (line *Line) removeString(count int, cursor *Location) error {
	if count == 0 {
		return nil
	}

	if !line.isValidLocation(cursor.getCol()) {
		return fmt.Errorf("[LINE ERROR] invalid cursor position, failed to remove string")
	}

	// This will be handled by the editor itself
	if cursor.getCol()-count < 0 {
		return fmt.Errorf("[LINE ERROR] invalid string length, failed to remove string")
	}

	line.content = line.content[0:cursor.getCol()-count] + line.content[cursor.getCol():]
	cursor.setCol(cursor.getCol() - count)
	return nil
}

func (line *Line) count() int {
	return len(line.content)
}

func (line *Line) Split(index int) (Line, Line) {
	first := newLine(line.content[:index])
	second := newLine(line.content[index:])
	return first, second
}

// get the start of the 'text' in the line 'line'
func (line *Line) search(startIndex int, text string) []int {
	var indices []int

	for i := startIndex; i < line.count()-len(text)+1; i++ {
		s := line.content[i : i+len(text)]
		if s == text {
			indices = append(indices, i)
		}
	}

	return indices
}

func (line *Line) replace(loc *Location, prevText, newText string) {
	col := loc.getCol()

	if !line.isValidLocation(col) {
		return
	}

	line.content = line.content[:col] + newText + line.content[col+len(prevText):]
	loc.setCol(col + len(newText))
}
