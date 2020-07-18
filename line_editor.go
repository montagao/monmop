package main

import (
	"regexp"
	"strings"

	"github.com/nsf/termbox-go"
)

type LineEditor struct {
	cmd        rune           // keyboard command such as "d" or "a"
	cursor     int            // cursor pos
	quoteIndex int            // currently highlighted quote
	prompt     string         // prompt string for a command
	input      string         // user typed input stirng
	quotes     *[]Quote       // pointer to quotes
	profile    *profile       // pointer to profile
	regex      *regexp.Regexp // regex to split comma-delimited input string
	commandWin *Win
}

func NewLineEditor(profile *profile, quotes *[]Quote, commandWin *Win) *LineEditor {
	return &LineEditor{
		quotes:     quotes,
		profile:    profile,
		commandWin: commandWin,
	}
}

func (editor *LineEditor) Prompt(cmd rune, quoteIndex int) {
	fg, bg := termbox.ColorDefault, termbox.ColorDefault
	prompts := map[rune]string{
		'a': `add tickers: `,
		'd': `delete selected ticker? y/n :`,
		'/': `/`,
	}

	if prompt, ok := prompts[cmd]; ok {
		editor.prompt = prompt
		editor.cmd = cmd

		editor.commandWin.print(0, 0, fg, bg, editor.prompt)
		termbox.SetCursor(len(editor.prompt), editor.commandWin.y)
		termbox.Flush()
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
	defer termbox.Flush()
	termbox.HideCursor()
	editor.commandWin.Clear()
	editor.prompt = ""
	editor.input = ""
	editor.cursor = 0
}

func (editor *LineEditor) Execute(selectedQuote int) {
	switch editor.cmd {
	case 'a':
		tickers := editor.tokenize()
		if len(tickers) > 0 {
			editor.profile.Tickers = append(editor.profile.Tickers, tickers...)
		}
	case 'd':
		if strings.TrimSpace(strings.ToLower(editor.input)) == "y" {
			for id, _ := range *editor.quotes {
				if q := (*editor.quotes)[id]; id == selectedQuote {
					// remove from list of tickers
					editor.profile.Tickers = removeTicker(editor.profile.Tickers, q.Ticker)
					// TODO: prevent adding duplicate tickers
					break
				}
			}
		}
	case '/':
		// perform a search on a ticker

	}
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
	fg, bg := termbox.ColorDefault, termbox.ColorDefault
	if editor.cursor < len(editor.input) {
		// Insert the character in the middle of the input string.
		editor.input = editor.input[0:editor.cursor] + string(ch) + editor.input[editor.cursor:len(editor.input)]
	} else {
		// Append the character to the end of the input string.
		editor.input += string(ch)
	}
	// TODO: make a heplper function for printing
	editor.commandWin.print(len(editor.prompt), 0, fg, bg, editor.input)
	editor.moveRight()

}

func (editor *LineEditor) deletePrevChar() {
	fg, bg := termbox.ColorDefault, termbox.ColorDefault
	if editor.cursor > 0 {
		if editor.cursor < len(editor.input) {
			// remove character in the middle
			editor.input = editor.input[0:editor.cursor-1] + editor.input[editor.cursor:len(editor.input)]
		} else {
			// remove last input characters
			editor.input = editor.input[:len(editor.input)-1]
		}
		editor.commandWin.print(len(editor.prompt), 0, fg, bg, editor.input+` `)
		editor.moveLeft()
	}
}

func (editor *LineEditor) moveRight() {
	if editor.cursor < len(editor.input) {
		editor.cursor++
		termbox.SetCursor(len(editor.prompt)+editor.cursor, editor.commandWin.y)
	}
}

func (editor *LineEditor) moveLeft() {
	if editor.cursor > 0 {
		editor.cursor--
		termbox.SetCursor(len(editor.prompt)+editor.cursor, editor.commandWin.y)
	}
}

func (editor *LineEditor) tokenize() []string {
	fields := strings.Split(strings.ToUpper(editor.input), ",")
	for i := range fields {
		fields[i] = strings.TrimSpace(fields[i])
	}
	return fields
}
