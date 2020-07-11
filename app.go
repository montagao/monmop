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
	ui := newUI()

	user, err := user.Current()
	if err != nil {
		panic(err)
	}

	profile, err := loadProfile(user)
	if err != nil {
		panic(err)
	}

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
		profile:  profile,
	}

}

func (app *app) GetQuotes() *[]Quote {
	quotes, err := FetchAll(app.profile.Tickers)

	if err != nil {
		panic(err)
	}

	return &quotes
}

// main app loop
func (app *app) loop() {
	quotes := app.GetQuotes()
	app.ui.draw(quotes)
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
			quotes := app.GetQuotes()
			app.ui.draw(quotes)
		}
	}
}
