package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/signal"
	"os/user"
	"path"
	"syscall"
	"time"

	"github.com/nsf/termbox-go"
)

const defaultProfile = ".config/monmop/"
const DEFAULT_DEBOUNCE_DURATION = 100 * time.Millisecond

type mode int

var file *os.File

const (
	NORMAL mode = iota
	COMMAND
	SORT
	CONFIRM_QUIT // new mode for quit confirmation
)

var navBindingKeys = map[termbox.Key]rune{
	termbox.KeyArrowDown:  'j',
	termbox.KeyArrowUp:    'k',
	termbox.KeyArrowLeft:  'h',
	termbox.KeyArrowRight: 'l',
}

var navKeys = map[rune]int{
	'j': 6,
	'k': 7,
	'h': 0,
	'l': 1,
	'b': 2,
	'e': 3,
	'0': 4,
	'$': 5,
}

type app struct {
	ui       *Ui
	ticker   *time.Ticker
	quitChan chan bool
	keyQueue chan termbox.Event
	profile  *profile
	mode     *mode

	// debounce keypresses
	allowOpenInBrowser bool
	debounceDuration   time.Duration
}

type portfolio struct {
	Tickers []string // list of stock tickers to display
}

type profile struct {
	Portfolios map[string]portfolio
	filepath   string
	Tickers    []string
}

func (profile *profile) Save() error {
	data, err := json.Marshal(profile)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(profile.filepath, data, 0644)
}

func loadProfile(user *user.User) (*profile, error) {
	profile := &profile{}

	profilePath := path.Join(user.HomeDir, defaultProfile)
	profile.filepath = profilePath

	if _, err := os.Stat(profile.filepath); os.IsNotExist(err) {
		os.Mkdir(profile.filepath, 0700)
	}

	profile.filepath = path.Join(profilePath, "monmoprc")
	data, err := ioutil.ReadFile(profile.filepath)
	if err != nil {
		// set some defaults
		profile.Portfolios = map[string]portfolio{
			"default": {
				Tickers: []string{"GOOG", "AAPL", "AMZN", "MSFT", "SPLK"},
			},
		}
		profile.Tickers = profile.Portfolios["default"].Tickers
		profile.Save()
	} else {
		json.Unmarshal(data, profile)
	}
	profile.Tickers = make([]string, len(profile.Portfolios["default"].Tickers))
	copy(profile.Tickers, profile.Portfolios["default"].Tickers)

	return profile, nil
}

func (app *app) saveProfile() error {
	b, err := json.Marshal(app.profile)
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(app.profile.filepath, b, 0644)
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

	signal.Notify(osChan,
		os.Interrupt,
		syscall.SIGHUP,
		syscall.SIGTERM,
		syscall.SIGABRT,
		syscall.SIGINT,
	)

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
		ui:                 ui,
		quitChan:           quitChan,
		ticker:             time.NewTicker(60 * time.Second),
		keyQueue:           keyQueue,
		profile:            profile,
		mode:               &mode,
		allowOpenInBrowser: true,
		debounceDuration:   DEFAULT_DEBOUNCE_DURATION,
	}

}

// main app loop
func (app *app) loop() {
	app.fetchAndDraw()
	defer file.Close()
	for {
		select {
		case <-app.quitChan:
			// TODO: disable this until we can handle this better,
			// i.e. with some kind of user prompt
			app.saveProfile()
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
						app.ui.Draw()
					} else {
						app.ui.HandleLineEditorInput(event)
					}
				case NORMAL:
					if event.Ch == 'q' || event.Ch == 'Q' {
						app.saveProfile()
						return
					} else if event.Ch == 'j' || event.Key == termbox.KeyArrowDown {
						app.ui.navigateStockDown()
						app.ui.Draw()
					} else if event.Ch == 'k' || event.Key == termbox.KeyArrowUp {
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
					} else if event.Ch == ':' {
						// a for "add"
						app.ui.Prompt(event.Ch)
						*app.mode = COMMAND
					} else if event.Ch == '/' {
						// a for "add"
						app.ui.Prompt(event.Ch)
						*app.mode = COMMAND
					} else if event.Ch == 'o' || event.Key == termbox.KeyEnter {
						// Check if OpenInBrowser action is allowed
						if app.allowOpenInBrowser {
							app.ui.OpenInBrowser()

							// Debounce the action
							app.allowOpenInBrowser = false
							time.AfterFunc(app.debounceDuration, func() {
								app.allowOpenInBrowser = true
							})
						}
					}
				case SORT:
					if event.Ch == 'q' || event.Ch == 'Q' {
						// app.saveProfile()
						return
					} else if isLabelNavigationEvent(event) {
						if ch, ok := navBindingKeys[event.Key]; ok {
							app.ui.HandleSortEvent(ch)
						} else {
							app.ui.HandleSortEvent(event.Ch)
						}
					} else if event.Key == termbox.KeyEsc {
						*app.mode = NORMAL
						app.ui.Draw()
					}
				}

			case termbox.EventResize:
				app.ui.Resize()
			}
		case <-app.ticker.C:
			app.fetchAndDraw()
		}
	}
}
func isLabelNavigationEvent(e termbox.Event) bool {
	if _, ok := navKeys[e.Ch]; ok {
		return true
	}
	return false
}

func (app *app) fetchAndDraw() {
	app.ui.GetQuotes()
	app.ui.Draw()
}
