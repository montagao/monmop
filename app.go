package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"os/user"
	"path"
	"syscall"
	"time"

	"github.com/nsf/termbox-go"
)

// TODO: make this more system-wide compatible
const defaultProfile = ".config/monmop/monmoprc"

type mode int

const (
	NORMAL mode = iota
	COMMAND
)

type app struct {
	ui       *Ui
	ticker   *time.Ticker
	quitChan chan bool
	keyQueue chan termbox.Event
	profile  *profile
}

type profile struct {
	Tickers  []string // list of stock tickers to display
	filepath string
}

func (profile *profile) Save() error {
	data, err := json.Marshal(profile)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(profile.filepath, data, 0644)
}

func loadProfile(user *user.User) (*profile, error) {
	// TODO: allow custom profile path? nah.
	profile := &profile{}

	// TODO: create dir if doens't exist
	profilePath := path.Join(user.HomeDir, defaultProfile)
	profile.filepath = profilePath

	data, err := ioutil.ReadFile(profilePath)
	if err != nil {
		// set some defaults
		profile.Tickers = []string{"CASH", "SPLK", "GOOG"}
		profile.Save()
	} else {
		json.Unmarshal(data, profile)
	}

	return profile, nil
}

func newApp() *app {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}

	profile, err := loadProfile(user)
	if err != nil {
		panic(err)
	}

	ui := newUI(profile)

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
		ticker:   time.NewTicker(60 * time.Second),
		keyQueue: keyQueue,
		profile:  profile,
	}

}

// main app loop
func (app *app) loop() {
	app.fetchAndDraw()
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
				} else if event.Ch == 'j' {
					app.ui.navigateStockDown()
					app.ui.draw()
				} else if event.Ch == 'k' {
					app.ui.navigateStockUp()
					app.ui.draw()
				} else if event.Ch == 'r' {
					// r for  "refresh"
					app.fetchAndDraw()
				} else if event.Ch == 'a' {
					// a for "add"

				}
			}
		case <-app.ticker.C:
			app.fetchAndDraw()
		}
	}
}

func (app *app) fetchAndDraw() {
	app.ui.GetQuotes()
	app.ui.draw()
}
