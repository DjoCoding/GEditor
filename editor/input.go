package editor

func (editor *Editor) setInputBufferInputRequestString(req string) {
	editor.input.req = req
}

func (editor *Editor) enableInputBuffer() {
	editor.input.enabled = true
}

func (editor *Editor) disableInputBuffer() {
	editor.input.enabled = false
}

func (editor *Editor) inputBufferIsEnabled() bool {
	return editor.input.enabled
}

func (editor *Editor) insertCharToInputBuffer(c rune) {
	if !editor.inputBufferIsEnabled() {
		return
	}

	editor.input.buffers[editor.input.current] += string(c)
}

func (editor *Editor) removeCharFromInputBuffer() {
	if !editor.inputBufferIsEnabled() {
		return
	}

	content := editor.input.buffers[editor.input.current]
	content = content[:len(content)-1]
	editor.input.buffers[editor.input.current] = content
}

func (editor *Editor) setInputCurrentBuffer(current int) {
	editor.input.current = current
}

func (editor *Editor) getInputCurrentBuffer() int {
	return editor.input.current
}

func (editor *Editor) resetInput() {
	editor.input = EditorInternalInput{}
}
