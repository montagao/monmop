package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nsf/termbox-go"
)

type app struct {
	ui       *Ui
	ticker   *time.Ticker
	quitChan chan bool
	keyQueue chan termbox.Event
}

func newApp() *app {
	ui := newUI()

	quitChan := make(chan bool, 1)
	osChan := make(chan os.Signal, 1)
	keyQueue := make(chan termbox.Event)

	signal.Notify(osChan, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM)
	go func() {
		<-osChan
		// quit on any OS kill/interrupt signal
		quitChan <- true
		return
	}()

	go func() {
		for {
			keyQueue <- termbox.PollEvent()
		}
	}()

	return &app{
		ui:       ui,
		quitChan: quitChan,
		ticker:   time.NewTicker(1 * time.Second),
		keyQueue: keyQueue,
	}

}

// main app loop
func (app *app) loop() {
	app.ui.draw()
	for {
		select {
		case <-app.quitChan:
			// TODO: save config on quit
			fmt.Printf("bye!")
			return // exit app
		case event := <-app.keyQueue:
			switch event.Type {
			case termbox.EventKey:
				if event.Ch == 'q' || event.Ch == 'Q' {
					fmt.Printf("See ya!")
					return
				}
			}
		case <-app.ticker.C:
			app.ui.draw()
		}
	}
}
