package editor

import "fmt"

type Line struct {
	content string
}

func NewLine(content string) Line {
	return Line{
		content: content,
	}
}

func (line *Line) GetContent() string {
	return line.content
}

func (line *Line) isValidLocation(cursor int) bool {
	return cursor <= line.Count()
}

func (line *Line) InsertString(s string, cursor *Location) error {
	if !line.isValidLocation(cursor.GetCol()) {
		return fmt.Errorf("[LINE ERROR] invalid cursor position, failed to append string %s", s)
	}

	line.content = line.content[0:cursor.GetCol()] + s + line.content[cursor.GetCol():]
	cursor.SetCol(cursor.GetCol() + len(s))
	return nil
}

func (line *Line) InsertChar(c rune, cursor *Location) error {
	return line.InsertString(string(c), cursor)
}

func (line *Line) RemoveString(count int, cursor *Location) error {
	if count == 0 {
		return nil
	}

	if !line.isValidLocation(cursor.GetCol()) {
		return fmt.Errorf("[LINE ERROR] invalid cursor position, failed to remove string")
	}

	// This will be handled by the editor itself
	if cursor.GetCol()-count < 0 {
		return fmt.Errorf("[LINE ERROR] invalid string length, failed to remove string")
	}

	line.content = line.content[0:cursor.GetCol()-count] + line.content[cursor.GetCol():]
	cursor.SetCol(cursor.GetCol() - count)
	return nil
}

func (line *Line) RemoveChar(cursor *Location) error {
	return line.RemoveString(1, cursor)
}

func (line *Line) Count() int {
	return len(line.content)
}

func (line *Line) Split(index int) (Line, Line) {
	first := NewLine(line.content[:index])
	second := NewLine(line.content[index:])
	return first, second
}

// get the start of the 'text' in the line 'line'
func (line *Line) Search(startIndex int, text string) []int {
	var indices []int

	for i := startIndex; i < line.Count()-len(text)+1; i++ {
		s := line.content[i : i+len(text)]
		if s == text {
			indices = append(indices, i)
		}
	}

	return indices
}
