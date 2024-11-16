package editor

type Location struct {
	line int
	col  int
}

func NewLocation(line, col int) Location {
	return Location{
		line: line,
		col:  col,
	}
}

func (cursor *Location) SetLine(line int) {
	cursor.line = line
}

func (cursor *Location) SetCol(col int) {
	cursor.col = col
}

func (cursor *Location) GetLine() int {
	return cursor.line
}

func (cursor *Location) GetCol() int {
	return cursor.col
}

func (cursor *Location) Set(line int, col int) {
	cursor.SetLine(line)
	cursor.SetCol(col)
}

func (cursor *Location) Get() (line int, col int) {
	return cursor.GetLine(), cursor.GetCol()
}

func (cursor *Location) Cmp(cur Location) bool {
	return cursor.GetLine() == cur.GetLine() && cursor.GetCol() == cur.GetCol()
}
