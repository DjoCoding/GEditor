package editor

import (
	"fmt"
	"os"
)

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

func (editor *Editor) checkFileInfoAndGetFile(fileInfo os.FileInfo) (*os.File, error) {
	if fileInfo.IsDir() {
		return nil, fmt.Errorf("can not handle directories right now")
	}

	return os.OpenFile(editor.config.CurrentFile, os.O_WRONLY|os.O_TRUNC, fileInfo.Mode().Perm())
}

func (editor *Editor) getInputFile() (*os.File, error) {
	fileInfo, err := os.Stat(editor.config.CurrentFile)
	if err == nil {
		return editor.checkFileInfoAndGetFile(fileInfo)
	}

	if os.IsNotExist(err) {
		return os.OpenFile(editor.config.CurrentFile, os.O_CREATE|os.O_WRONLY, 0644)
	}

	panic("unhandled situation")
}

// save into a hardcoded filepath
func (editor *Editor) save() error {
	f, err := editor.getInputFile()
	if err != nil {
		return err
	}

	err = editor.saveContent(f)
	if err != nil {
		return err
	}

	return f.Close()
}
