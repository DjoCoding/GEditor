package editor

import "fmt"

const (
	BUFFER_INITIAL_CAPACITY = 1
	BUFFER_TAB_SIZE         = 4
)

type Buffer struct {
	lines []Line
}

func NewBuffer() Buffer {
	return Buffer{
		lines: make([]Line, BUFFER_INITIAL_CAPACITY),
	}
}

func (buffer *Buffer) isValidLine(lineIndex int) bool {
	return lineIndex < len(buffer.lines)
}

func (buffer *Buffer) isEmpty() bool {
	return len(buffer.lines) == 1 && buffer.lines[0].Count() == 0
}

func (buffer *Buffer) LastLineCount() int {
	return buffer.lines[buffer.Count()-1].Count()
}

func (buffer *Buffer) Count() int {
	return len(buffer.lines)
}

func (buffer *Buffer) InsertString(s string, cursor *Location) error {
	if !buffer.isValidLine(cursor.GetLine()) {
		return fmt.Errorf("[BUFFER ERROR] invalid cursor position, failed to append string")
	}

	return buffer.lines[cursor.GetLine()].InsertString(s, cursor)
}

func (buffer *Buffer) InsertChar(c rune, cursor *Location) error {
	return buffer.InsertString(string(c), cursor)
}

func (buffer *Buffer) RemoveLine(lineIndex int) error {
	if !buffer.isValidLine(lineIndex) {
		return fmt.Errorf("[BUFFER ERROR] invalid line index, faild to remove line")
	}

	lines := make([]Line, 0, buffer.Count()-1)
	for row, line := range buffer.lines {
		if row == lineIndex {
			continue
		}
		lines = append(lines, line)
	}

	buffer.lines = lines

	return nil
}

func (buffer *Buffer) AppendLineContent(lineIndex, lineIndexToAppend int) {
	buffer.lines[lineIndex].content += buffer.lines[lineIndexToAppend].content
}

func (buffer *Buffer) RemoveString(count int, cursor *Location) error {
	line, col := cursor.Get()
	if line == 0 && col == 0 {
		return nil
	}

	if buffer.isEmpty() {
		return nil
	}

	if count == 0 {
		return nil
	}

	if !buffer.isValidLine(cursor.GetLine()) {
		return fmt.Errorf("[BUFFER ERROR] invalid cursor position, failed to remove string")
	}

	charsCount := cursor.GetCol()
	if charsCount > count {
		charsCount = count
	}

	err := buffer.lines[cursor.GetLine()].RemoveString(charsCount, cursor)
	if err != nil {
		return err
	}

	count -= charsCount

	switch {
	case buffer.isEmpty():
		return nil

	case count >= 1:
		prevLineCount := buffer.lines[cursor.GetLine()-1].Count()

		buffer.AppendLineContent(cursor.GetLine()-1, cursor.GetLine())
		buffer.RemoveLine(cursor.GetLine())
		count -= 1

		// set the cursor to its place
		cursor.SetLine(cursor.GetLine() - 1)
		cursor.SetCol(prevLineCount)
	}

	return buffer.RemoveString(count, cursor)
}

func (buffer *Buffer) RemoveChar(cursor *Location) error {
	if cursor.GetCol() < BUFFER_TAB_SIZE {
		return buffer.RemoveString(1, cursor)
	}

	isTab := true
	for i := cursor.GetCol() - BUFFER_TAB_SIZE; i < cursor.GetCol(); i++ {
		if buffer.lines[cursor.GetLine()].content[i] != ' ' {
			isTab = false
			break
		}
	}

	if isTab {
		return buffer.RemoveString(BUFFER_TAB_SIZE, cursor)
	}

	return buffer.RemoveString(1, cursor)
}

func (buffer *Buffer) InsertNewLine(cursor *Location) error {
	if !buffer.isValidLine(cursor.GetLine()) {
		return fmt.Errorf("[BUFFER ERROR] invalid cursor position, failed to insert a new line")
	}

	lines := make([]Line, 0, buffer.Count()+1)
	for row, line := range buffer.lines {
		if row == cursor.GetLine() {
			up, down := line.Split(cursor.GetCol())
			lines = append(lines, up)
			lines = append(lines, down)
			continue
		}
		lines = append(lines, line)
	}

	buffer.lines = lines

	cursor.SetLine(cursor.GetLine() + 1)
	cursor.SetCol(0)

	return nil
}

func (buffer *Buffer) InsertTab(cursor *Location) error {
	for i := 0; i < BUFFER_TAB_SIZE; i++ {
		err := buffer.InsertChar(' ', cursor)
		if err != nil {
			return err
		}
	}

	return nil
}

// get the end of the 'text' in the whole buffer
func (buffer *Buffer) Search(currentLocation Location, text string) (Location, bool) {
	// search in current row
	col := 0
	found := false

	col, found = buffer.lines[currentLocation.GetLine()].Search(currentLocation.GetCol(), text)
	if found {
		return Location{
			line: currentLocation.GetLine(),
			col:  col,
		}, true
	}

	for i := currentLocation.GetLine() + 1; i < buffer.Count(); i++ {
		col, found = buffer.lines[i].Search(0, text)
		if found {
			return Location{
				line: i,
				col:  col,
			}, true
		}
	}

	for i := 0; i < currentLocation.GetLine(); i++ {
		col, found := buffer.lines[i].Search(0, text)
		if found {
			return Location{
				line: i,
				col:  col,
			}, true
		}
	}

	col, found = buffer.lines[currentLocation.GetLine()].Search(0, text)
	if !found {
		return NewLocation(), false
	}

	return Location{
		line: currentLocation.GetLine(),
		col:  col,
	}, true
}
