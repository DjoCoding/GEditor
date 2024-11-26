package editor

type Location struct {
	line int
	col  int
}

func newLocation(line, col int) Location {
	return Location{
		line: line,
		col:  col,
	}
}

func (cursor *Location) setLine(line int) {
	cursor.line = line
}

func (cursor *Location) setCol(col int) {
	cursor.col = col
}

func (cursor *Location) getLine() int {
	return cursor.line
}

func (cursor *Location) getCol() int {
	return cursor.col
}

func (cursor *Location) set(line int, col int) {
	cursor.setLine(line)
	cursor.setCol(col)
}

func (cursor *Location) get() (line int, col int) {
	return cursor.getLine(), cursor.getCol()
}

func (cursor *Location) cmp(cur Location) bool {
	return cursor.getLine() == cur.getLine() && cursor.getCol() == cur.getCol()
}
