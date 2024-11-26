package editor

import (
	"fmt"
	"os"
)

// load a file using the EditorConfiguration fields (passed as args)
func (editor *Editor) loadFileFromConfiguration() error {
	fileInfo, err := os.Stat(editor.config.Filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if fileInfo.IsDir() {
		return fmt.Errorf("can not open directories in this text editor")
	}

	fileContent, err := os.ReadFile(editor.config.Filepath)
	if err != nil {
		return err
	}

	for _, c := range fileContent {
		switch c {
		case '\n':
			err = editor.insertNewLine()
		case '\t':
			err = editor.insertTab()
		default:
			err = editor.loadCharFromFile(rune(c))
		}

		if err != nil {
			return err
		}
	}

	return nil
}

// load a char from a file into the editor buffer
func (editor *Editor) loadCharFromFile(c rune) error {
	return editor.buffer.insertCharNormally(c, &editor.realCursor)
}

// load file to the editor buffer
// main function
func (editor *Editor) Load() error {
	if editor.config.Filepath == "" {
		return nil
	}

	return editor.loadFileFromConfiguration()
}
