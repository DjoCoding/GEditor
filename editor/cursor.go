package editor

type Cursor struct {
	line int
	col  int
}

func NewCursor() Cursor {
	return Cursor{
		line: 0,
		col:  0,
	}
}

func (cursor *Cursor) SetLine(line int) {
	cursor.line = line
}

func (cursor *Cursor) SetCol(col int) {
	cursor.col = col
}

func (cursor *Cursor) GetLine() int {
	return cursor.line
}

func (cursor *Cursor) GetCol() int {
	return cursor.col
}

func (cursor *Cursor) Set(line int, col int) {
	cursor.SetLine(line)
	cursor.SetCol(col)
}

func (cursor *Cursor) Get() (line int, col int) {
	return cursor.GetLine(), cursor.GetCol()
}
