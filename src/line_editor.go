package main

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/nsf/termbox-go"
)

type LineEditor struct {
	cmd         rune   // keyboard command such as "d" or "a"
	cursor      int    // cursor pos
	quoteIndex  int    // currently highlighted quote
	prompt      string // prompt string for a command
	input       string // user typed input string
	promptError string
	message     string
	quotes      *[]Quote       // pointer to quotes
	profile     *profile       // pointer to profile
	regex       *regexp.Regexp // regex to split comma-delimited input string
	commandWin  *Win
}

func NewLineEditor(profile *profile, quotes *[]Quote, commandWin *Win) *LineEditor {
	return &LineEditor{
		quotes:     quotes,
		profile:    profile,
		commandWin: commandWin,
	}
}

func (editor *LineEditor) Prompt(cmd rune, quoteIndex int) {
	prompts := map[rune]string{
		'a': `add tickers: `,
		'd': `delete selected ticker? y/n :`,
		'/': `/`,
		':': `:`,
	}

	if prompt, ok := prompts[cmd]; ok {
		editor.prompt = prompt
		editor.cmd = cmd
	}
}

func (editor *LineEditor) Draw() {
	fg, bg := termbox.ColorDefault, termbox.ColorDefault
	editor.commandWin.Clear()
	if editor.promptError != "" {
		fg, bg = termbox.ColorRed, termbox.ColorWhite
		editor.commandWin.print(0, 0, fg, bg, editor.promptError)
		editor.Done()
		return
	} else if editor.message != "" {
		editor.commandWin.print(0, 0, fg, bg, editor.message)
		editor.Done()
		return
	} else if editor.prompt != "" {
		editor.commandWin.print(0, 0, fg, bg, editor.prompt)
		editor.commandWin.print(len(editor.prompt), 0, fg, bg, editor.input)
		termbox.SetCursor(len(editor.prompt)+len(editor.input), editor.commandWin.y)
	}
}

func (editor *LineEditor) Handle(ev termbox.Event) {
	defer termbox.Flush()

	switch ev.Key {
	case termbox.KeySpace:
		editor.insertCharacter(' ')
	case termbox.KeyBackspace, termbox.KeyBackspace2:
		editor.deletePrevChar()
	case termbox.KeyArrowLeft:
		editor.moveLeft()
	case termbox.KeyArrowRight:
		editor.moveRight()
	default:
		if ev.Ch != 0 {
			editor.insertCharacter(ev.Ch)
		}
	}
}

func (editor *LineEditor) Done() {
	termbox.HideCursor()
	editor.prompt = ""
	editor.input = ""
	editor.message = ""
	editor.promptError = ""
	editor.cursor = 0

	if editor.prompt != "" {
		termbox.SetCursor(len(editor.prompt)+len(editor.input), editor.commandWin.y)
	}

}

func (editor *LineEditor) Execute(selectedQuote int) (newQuote int) {
	switch editor.cmd {
	case 'd':
		if strings.TrimSpace(strings.ToLower(editor.input)) == "y" {
			for id := range *editor.quotes {
				if q := (*editor.quotes)[id]; id == selectedQuote {
					// remove from list of tickers
					editor.profile.Tickers = removeTicker(editor.profile.Tickers, q.Ticker)
					return selectedQuote - 1
				}
			}
		}
	case '/':
		// perform a search on a ticker
		return getTickerId(editor.profile.Tickers, strings.TrimSpace(strings.ToUpper(editor.input)))
	case ':':
		args := editor.tokenize(" ")
		editor.input = ""
		termbox.HideCursor()
		termbox.Flush()
		if args[0] == "save" {
			portfolioName := args[1]
			editor.profile.Portfolios[portfolioName] = portfolio{
				Tickers: append([]string{}, editor.profile.Tickers...),
			}
			editor.message = fmt.Sprintf("saved portfolio as '%s'", portfolioName)
		} else if args[0] == "load" {
			portfolioName := args[1]
			portfolio, ok := editor.profile.Portfolios[portfolioName]
			if !ok {
				editor.PrintErrorf("portfolio not found: %s", portfolioName)
				return -1
			} else {
				editor.profile.Tickers = append([]string{}, portfolio.Tickers...)

				editor.message = fmt.Sprintf("loaded portfolio '%s'", portfolioName)
			}
		} else if args[0] == "new" {
			editor.profile.Tickers = []string{}
			editor.message = fmt.Sprintf("creating new portfolio")
		} else if args[0] == "list" {
			editor.message = fmt.Sprintf("saved portfolios: '%s'", reflect.ValueOf(editor.profile.Portfolios).MapKeys())
		} else {
			editor.PrintErrorf("could not recognize command '%s'", args[0])
		}
		return 0
	}
	return 0
}

func (editor *LineEditor) PrintErrorf(format string, a ...interface{}) {
	editor.promptError = fmt.Sprintf(format, a...)
}

func (editor *LineEditor) AddQuotes() (ticker string, err error) {
	// input for tickers should match "GOOG" or "GOOG, AAPL" ignoring whitespace inbetween
	validFmt := regexp.MustCompile(`^\s*([a-zA-Z-]+)(,\s*[a-zA-Z-]+\s*)*$`)
	if !validFmt.MatchString(editor.input) {
		return editor.input, fmt.Errorf("parse error")
	}
	tickers := editor.tokenize(",")
	for _, ticker := range tickers {
		// ignore duplicate tickers
		if getTickerId(editor.profile.Tickers, ticker) == -1 {
			editor.profile.Tickers = append(editor.profile.Tickers, ticker)
		}
	}

	// return the last ticker added so we can select it
	return tickers[len(tickers)-1], nil
}

func removeTicker(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}
func (editor *LineEditor) insertCharacter(ch rune) {
	if editor.cursor < len(editor.input) {
		// Insert the character in the middle of the input string.
		editor.input = editor.input[0:editor.cursor] + string(ch) + editor.input[editor.cursor:len(editor.input)]
	} else {
		// Append the character to the end of the input string.
		editor.input += string(ch)
	}
	editor.moveRight()
}

func (editor *LineEditor) deletePrevChar() {
	if editor.cursor > 0 {
		if editor.cursor < len(editor.input) {
			// remove character in the middle
			editor.input = editor.input[0:editor.cursor-1] + editor.input[editor.cursor:len(editor.input)]
		} else {
			// remove last input characters
			editor.input = editor.input[:len(editor.input)-1]
		}
		editor.moveLeft()
	}
}

func (editor *LineEditor) moveRight() {
	if editor.cursor < len(editor.input) {
		editor.cursor++
	}
}

func (editor *LineEditor) moveLeft() {
	if editor.cursor > 0 {
		editor.cursor--
	}
}

func (editor *LineEditor) tokenize(delim string) []string {
	fields := strings.Split(editor.input, delim)
	for i := range fields {
		fields[i] = strings.TrimSpace(fields[i])
	}
	return fields
}

func getTickerId(tickers []string, ticker string) int {
	for p, v := range tickers {
		if v == ticker {
			return p
		}
	}
	return -1
}
