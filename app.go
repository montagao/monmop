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
	SORT
)

type app struct {
	ui       *Ui
	ticker   *time.Ticker
	quitChan chan bool
	keyQueue chan termbox.Event
	profile  *profile
	mode     *mode
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

func (app *app) saveProfile() error {
	user, err := user.Current()
	if err != nil {
		return err
	}
	profilePath := path.Join(user.HomeDir, defaultProfile)

	var b []byte
	b, err = json.Marshal(app.profile)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(profilePath, b, 0644)
	if err != nil {
		return err
	}
	return nil
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

	mode := NORMAL
	ui := newUI(profile, &mode)

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
		mode:     &mode,
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
				switch *app.mode {
				case COMMAND:
					if event.Key == termbox.KeyEnter {
						app.ui.ExecuteCommand()
						// app.fetchAndDraw()
						*app.mode = NORMAL
					} else if event.Key == termbox.KeyEsc {
						app.ui.lineEditor.Done()
						*app.mode = NORMAL
					} else {
						app.ui.HandleInputKey(event)
					}
				case NORMAL:
					if event.Ch == 'q' || event.Ch == 'Q' {
						app.saveProfile()
						return
					} else if event.Ch == 'j' {
						app.ui.navigateStockDown()
						app.ui.Draw()
					} else if event.Ch == 'k' {
						app.ui.navigateStockUp()
						app.ui.Draw()
					} else if event.Ch == 'g' {
						app.ui.navigateStockBeginning()
						app.ui.Draw()
					} else if event.Ch == 'G' {
						app.ui.navigateStockEnd()
						app.ui.Draw()
					} else if event.Ch == 'r' {
						// r for  "refresh"
						app.fetchAndDraw()
					} else if event.Ch == 's' {
						// s for  "sort"
						*app.mode = SORT
						app.ui.Draw()
					} else if event.Ch == 'a' {
						// a for "add"
						app.ui.Prompt(event.Ch)
						*app.mode = COMMAND
					} else if event.Ch == 'd' {
						// a for "add"
						app.ui.Prompt(event.Ch)
						*app.mode = COMMAND
					} else if event.Ch == '/' {
						// a for "add"
						app.ui.Prompt(event.Ch)
						*app.mode = COMMAND
					} else if event.Ch == 'o' || event.Key == termbox.KeyEnter {
						app.ui.OpenInBrowser()
					}
				case SORT:
					if event.Ch == 'q' || event.Ch == 'Q' {
						app.saveProfile()
						return
					} else if event.Ch == 'h' || event.Ch == 'l' || event.Ch == 'b' || event.Ch == 'e' {
						app.ui.NavigateLabel(event.Ch)
					} else if event.Ch == '0' || event.Ch == '$' {
						app.ui.NavigateLabel(event.Ch)
					} else if event.Ch == 'j' || event.Ch == 'k' {
						app.ui.SortLabel(event.Ch)
					} else if event.Key == termbox.KeyEsc {
						*app.mode = NORMAL
						app.ui.Draw()
					}
				}

			case termbox.EventResize:
				termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
				app.ui.Resize()
				app.ui.Draw()
			}
		case <-app.ticker.C:
			app.fetchAndDraw()
		}
	}
}

func (app *app) fetchAndDraw() {
	app.ui.GetQuotes()
	app.ui.Draw()
}
