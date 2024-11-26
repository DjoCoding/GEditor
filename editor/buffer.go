package editor

import "fmt"

const (
    BUFFER_INITIAL_CAPACITY = 1
    BUFFER_TAB_SIZE         = 4
)

type Buffer struct {
    lines []Line
}

func newBuffer() Buffer {
    return Buffer{
        lines: make([]Line, BUFFER_INITIAL_CAPACITY),
    }
}

func (buffer *Buffer) isValidLine(lineIndex int) bool {
    return lineIndex < len(buffer.lines)
}

func (buffer *Buffer) isEmpty() bool {
    return len(buffer.lines) == 1 && buffer.lines[0].count() == 0
}

func (buffer *Buffer) lastLineCount() int {
    return buffer.lines[buffer.count()-1].count()
}

func (buffer *Buffer) count() int {
    return len(buffer.lines)
}

func (buffer *Buffer) insertString(s string, cursor *Location) error {
    if !buffer.isValidLine(cursor.getLine()) {
        return fmt.Errorf("[BUFFER ERROR] invalid cursor position, failed to append string")
    }

    return buffer.lines[cursor.getLine()].insertString(s, cursor)
}

func getComplementaryChar(c rune) (comc rune, hasOne bool) {
    comc = 0

    switch c {
    case '{':
        comc = '}'
    case '(':
        comc = ')'
    case '[':
        comc = ']'
    case '"':
        comc = '"'
    case '\'':
        comc = '\''
    case '<':
        comc = '>'
    }

    hasOne = comc != 0
    return comc, hasOne
}

// insert a char into the editor buffer without matching chars completion
func (buffer *Buffer) insertCharNormally(c rune, cursor *Location) error {
    return buffer.insertString(string(c), cursor)
}

func (buffer *Buffer) insertChar(c rune, cursor *Location) error {
    comc, hasOne := getComplementaryChar(c)

    if !hasOne {
        return buffer.insertString(string(c), cursor)
    }

    err := buffer.insertString(string(c)+string(comc), cursor)
    if err != nil {
        return err
    }

    cursor.setCol(cursor.getCol() - 1)
    return nil
}

func (buffer *Buffer) removeLine(lineIndex int) error {
    if !buffer.isValidLine(lineIndex) {
        return fmt.Errorf("[BUFFER ERROR] invalid line index, faild to remove line")
    }

    lines := make([]Line, 0, buffer.count()-1)
    for row, line := range buffer.lines {
        if row == lineIndex {
            continue
        }
        lines = append(lines, line)
    }

    buffer.lines = lines

    return nil
}

func (buffer *Buffer) appendLineContent(lineIndex, lineIndexToAppend int) {
    buffer.lines[lineIndex].content += buffer.lines[lineIndexToAppend].content
}

func (buffer *Buffer) removeString(count int, cursor *Location) error {
    line, col := cursor.get()
    if line == 0 && col == 0 {
        return nil
    }

    if buffer.isEmpty() {
        return nil
    }

    if count == 0 {
        return nil
    }

    if !buffer.isValidLine(cursor.getLine()) {
        return fmt.Errorf("[BUFFER ERROR] invalid cursor position, failed to remove string")
    }

    charscount := cursor.getCol()
    if charscount > count {
        charscount = count
    }

    err := buffer.lines[cursor.getLine()].removeString(charscount, cursor)
    if err != nil {
        return err
    }

    count -= charscount

    switch {
    case buffer.isEmpty():
        return nil

    case count >= 1:
        prevLinecount := buffer.lines[cursor.getLine()-1].count()

        buffer.appendLineContent(cursor.getLine()-1, cursor.getLine())
        buffer.removeLine(cursor.getLine())
        count -= 1

        // set the cursor to its place
        cursor.setLine(cursor.getLine() - 1)
        cursor.setCol(prevLinecount)
    }

    return buffer.removeString(count, cursor)
}

func (buffer *Buffer) hasMatchingChars(cursor *Location) bool {
    line := buffer.lines[cursor.getLine()]

    if line.count()-cursor.getCol() >= 1 && cursor.getCol() > 0 {
        charAt := line.content[cursor.getCol()]
        charToRemove := line.content[cursor.getCol()-1]
        comc, hasOne := getComplementaryChar(rune(charToRemove))
        return hasOne && comc == rune(charAt)
    }

    return false
}

func (buffer *Buffer) removeMatchingChars(cursor *Location) error {
    cursor.setCol(cursor.getCol() + 1)
    return buffer.removeString(2, cursor)
}

func (buffer *Buffer) removeChar(cursor *Location) error {
    if cursor.getCol() < BUFFER_TAB_SIZE {
        if buffer.hasMatchingChars(cursor) {
            return buffer.removeMatchingChars(cursor)
        }

        return buffer.removeString(1, cursor)
    }

    isTab := true
    for i := cursor.getCol() - BUFFER_TAB_SIZE; i < cursor.getCol(); i++ {
        if buffer.lines[cursor.getLine()].content[i] != ' ' {
            isTab = false
            break
        }
    }

    if isTab {
        return buffer.removeString(BUFFER_TAB_SIZE, cursor)
    }

    if buffer.hasMatchingChars(cursor) {
        return buffer.removeMatchingChars(cursor)
    }

    return buffer.removeString(1, cursor)
}

func (buffer *Buffer) insertNewLine(cursor *Location) error {
    if !buffer.isValidLine(cursor.getLine()) {
        return fmt.Errorf("[BUFFER ERROR] invalid cursor position, failed to insert a new line")
    }

    lines := make([]Line, 0, buffer.count()+1)
    for row, line := range buffer.lines {
        if row == cursor.getLine() {
            up, down := line.Split(cursor.getCol())
            lines = append(lines, up)
            lines = append(lines, down)
            continue
        }
        lines = append(lines, line)
    }

    buffer.lines = lines

    cursor.setLine(cursor.getLine() + 1)
    cursor.setCol(0)

    return nil
}

func (buffer *Buffer) insertTab(cursor *Location) error {
    for i := 0; i < BUFFER_TAB_SIZE; i++ {
        err := buffer.insertChar(' ', cursor)
        if err != nil {
            return err
        }
    }

    return nil
}

// get the end of the 'text' in the whole buffer
func (buffer *Buffer) search(text string) []Location {
    var locations []Location

    for row, line := range buffer.lines {
        cols := line.search(0, text)
        for _, col := range cols {
            locations = append(locations, newLocation(row, col))
        }
    }

    return locations
}

func (buffer *Buffer) findAndReplace(newText, prevText string, location *Location) {
    if !buffer.isValidLine(location.getLine()) {
        return
    }
    buffer.lines[location.getLine()].replace(location, prevText, newText)
}

