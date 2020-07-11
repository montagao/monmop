package main

import (
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"
)

// win
type Win struct {
	w, h, x, y int
}

func (win *Win) print(x, y int, fg, bg termbox.Attribute, s string) (termbox.Attribute, termbox.Attribute) {

	for i := 0; i < len(s); i++ {
		// decodes a utf8 from a string
		// since strings in go are really just immutable []bytes
		r, w := utf8.DecodeRuneInString(s[i:])

		// TOOD: handle escape codes and control characters?

		if x < win.w {
			// fmt.Printf("printing %c\n", r)
			termbox.SetCell(win.x+x, win.y+y, r, fg, bg)
		}

		i += w - 1

		// TOOD: handle tabs

		x += runewidth.RuneWidth(r)
	}

	return fg, bg

}

// thin wrapper around TermBox to provide basic UI for monmop
type Ui struct {
	titleWin *Win
	// marketWin *Win
	labelWin *Win
	// stockWin  *Win
}

func newUI() *Ui {
	wtot, _ := termbox.Size()

	eventQ := make(chan termbox.Event)
	go func() {
		for {
			eventQ <- termbox.PollEvent()
		}
	}()

	eventChann := make(chan termbox.Event)
	go func() {
		for {
			e := <-eventQ
			// TODO
			// handle  alt modifiers?o
			// there's some more work to be done here.
			eventChann <- e
		}
	}()

	return &Ui{
		titleWin: &Win{
			w: wtot,
			h: 1,
			x: 0,
			y: 0,
		},
		labelWin: &Win{
			w: wtot,
			h: 1,
			x: 0,
			y: 1,
		},
	}

}

func (ui *Ui) draw() {
	fg, bg := termbox.ColorDefault, termbox.ColorDefault

	termbox.Clear(fg, bg)
	ui.drawTitleLine()
	ui.drawLabelWin()

	// w, h := termbox.Size()
	// fmt.Printf("termbox size: %d, %d", w, h)
	termbox.Flush()
}

// Temp for playing aroudn with termbox
func (ui *Ui) drawTitleLine() {
	fg, bg := termbox.ColorDefault, termbox.ColorDefault

	title := "Monmop, created by Monta"

	ui.titleWin.print(0, 0, fg, bg, title)
}

func (ui *Ui) drawLabelWin() {
	fg, bg := termbox.ColorDefault, termbox.ColorDefault

	// TODO: don't do this. use a struct with properties, or some constants.
	labels := "Ticker | Last | Change | Josh"

	ui.labelWin.print(0, 0, fg, bg, labels)
}
