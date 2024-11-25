package editor

import "os"

// save the content of the editor buffer to a file
// main function
func (editor *Editor) saveContent(f *os.File) error {
	for _, line := range editor.buffer.lines {
		_, err := f.Write([]byte(line.content))
		if err != nil {
			return err
		}
		_, err = f.Write([]byte("\n"))
		if err != nil {
			return err
		}
	}

	return nil
}

// not implemented yet
func (editor *Editor) saveFromConfiguration() error {
	return nil
}

// save into a hardcoded filepath
func (editor *Editor) save() error {
	if editor.config.Filepath != nil {
		return editor.saveFromConfiguration()
	}

	filepath := "./test"
	f, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	return editor.saveContent(f)
}
