package main

import (
    "edit/editor"
    "fmt"
    "os"
)

func main() {
    var config editor.EditorConfiguration

    // for now i can only pass the filepath as an argument
    for i := 1; i < len(os.Args); i++ {
        arg := os.Args[i]
        switch arg {
        default:
            config.OpenedFile = arg
        }
    }

    editor, err := editor.New(config)
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }

    err = editor.Load()
    if err != nil {
        editor.Close()
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

