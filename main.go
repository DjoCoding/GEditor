package main

import (
	"edit/editor"
	"fmt"
	"os"
)

func main() {
	editor, err := editor.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	defer editor.Close()

	for editor.ShouldNotQuit() {
		ev := editor.PollEvent()
		err = editor.HandleEvent(ev)
		if err != nil {
			editor.Quit()
		}

		editor.Render()
	}

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
