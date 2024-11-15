package editor

import "fmt"

const BUFFER_INITIAL_CAPACITY = 10

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

func (buffer *Buffer) InsertString(s string, cursor *Cursor) error {
	if !buffer.isValidLine(cursor.GetLine()) {
		return fmt.Errorf("[BUFFER ERROR] invalid cursor position, failed to append string")
	}

	return buffer.lines[cursor.GetLine()].InsertString(s, cursor)
}

func (buffer *Buffer) InsertChar(c rune, cursor *Cursor) error {
	return buffer.InsertString(string(c), cursor)
}

func (buffer *Buffer) RemoveLine(lineIndex int) error {
	if !buffer.isValidLine(lineIndex) {
		return fmt.Errorf("[BUFFER ERROR] invalid line index, faild to remove line")
	}

	var lines []Line = buffer.lines[:lineIndex]
	lines = append(lines, buffer.lines[lineIndex+1:]...)
	buffer.lines = lines

	return nil
}

func (buffer *Buffer) RemoveString(count int, cursor *Cursor) error {
	if buffer.isEmpty() {
		return nil
	}

	if count == 0 {
		return nil
	}

	if !buffer.isValidLine(cursor.GetLine()) {
		return fmt.Errorf("[BUFFER ERROR] invalid cursor position, failed to remove string")
	}

	charsCount := buffer.lines[cursor.GetLine()].Count()
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
		buffer.RemoveLine(cursor.GetLine())
		count -= 1

		// set the cursor to its place
		cursor.SetLine(cursor.GetLine() - 1)
		cursor.SetCol(buffer.lines[cursor.GetLine()].Count())
	}

	return buffer.RemoveString(count, cursor)
}

func (buffer *Buffer) RemoveChar(cursor *Cursor) error {
	return buffer.RemoveString(1, cursor)
}

// this will work only at the end of the line
func (buffer *Buffer) InsertNewLine(cursor *Cursor) error {
	if !buffer.isValidLine(cursor.GetLine()) {
		return fmt.Errorf("[BUFFER ERROR] invalid cursor position, failed to insert a new line")
	}

	var splittedLines [2]Line = buffer.lines[cursor.GetLine()].Split(cursor.GetCol())

	var lines []Line = buffer.lines[:cursor.GetLine()]
	lines = append(lines, splittedLines[0], splittedLines[1])
	lines = append(lines, buffer.lines[cursor.GetLine()+1:]...)

	buffer.lines = lines

	cursor.SetLine(cursor.GetLine() + 1)
	cursor.SetCol(0)

	return nil
}
