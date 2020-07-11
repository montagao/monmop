package main

import (
	"fmt"

	"github.com/nsf/termbox-go"
)

func main() {
	if err := termbox.Init(); err != nil {
		fmt.Printf("failed initializing termbox: %s", err)
	}

	defer termbox.Close()

	app := newApp()

	app.loop()

}
