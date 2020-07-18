package main

import (
	"fmt"
	"log"
	"os/exec"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"
)

var marketNames = map[string]string{
	"^DJI":     "Dow Jones",
	"^GSPC":    "S&P 500",
	"^IXIC":    "NASDAQ",
	"^N225":    "Nikkei",
	"^HSI":     "Hong Kong",
	"^FTSE":    "London",
	"^GDAXI":   "Frankfurt",
	"^TNX":     "10-Year Yield",
	"CAD=X":    "CAD",
	"EURUSD=X": "Euro",
	"RMB=F":    "RMB",
	"CL=F":     "Oil",
	"GC=F":     "Gold",
}

const (
	NO_CHAR         string = " "
	DESCENDING_CHAR string = "🠕"
	ASCENDING_CHAR  string = "🠗"
)

const appTitle = "monmop 0.1"

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

func (win *Win) Clear() {
	fg, bg := termbox.ColorDefault, termbox.ColorDefault
	for j := 0; j < win.h; j++ {
		for i := 0; i < win.w; i++ {
			termbox.SetCell(win.x+i, win.y+j, ' ', fg, bg)
		}
	}
}

// thin wrapper around TermBox to provide basic UI for monmop
type Ui struct {
	titleWin   *Win
	marketWin  *Win
	labelWin   *Win
	stockWin   *Win
	commandWin *Win

	layout *Layout

	selectedQuote        int
	zerothQuote          int
	selectedVisibleQuote int
	selectedSort         int
	stockQuotes          *[]Quote
	visibleQuotes        []Quote
	marketQuotes         *[]Quote
	maxQuotesHeight      int
	selectedLabel        int
	sortSymbol           string

	mode       *mode
	profile    *profile
	lineEditor *LineEditor
}

func newUI(profile *profile, mode *mode) *Ui {
	wtot, htot := termbox.Size()

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
		marketWin: &Win{
			w: wtot,
			h: 4,
			x: 0,
			y: 1,
		},
		labelWin: &Win{
			w: wtot,
			h: 1,
			x: 0,
			y: 5,
		},
		stockWin: &Win{
			w: wtot,
			h: htot - 7, //terrible practice, but idc
			x: 0,
			y: 6,
		},
		commandWin: &Win{
			w: wtot,
			h: 1,
			x: 0,
			y: htot - 1,
		},
		layout:          NewLayout(),
		selectedQuote:   0,
		selectedSort:    0,
		zerothQuote:     0,
		selectedLabel:   0,
		sortSymbol:      NO_CHAR,
		mode:            mode,
		profile:         profile,
		maxQuotesHeight: htot - 7,
		lineEditor: NewLineEditor(
			profile,
			nil,
			&Win{
				w: wtot,
				h: 1,
				x: 0,
				y: htot - 1,
			},
		),
	}

}

func (ui *Ui) Resize() {
	wtot, htot := termbox.Size()
	ui.titleWin.w = wtot
	ui.marketWin.w = wtot
	ui.labelWin.w = wtot
	ui.stockWin.w = wtot
	ui.stockWin.h = htot - 7
	ui.commandWin.w = wtot
	ui.commandWin.y = htot - 1
	ui.maxQuotesHeight = htot - 7
}

func (ui *Ui) Draw() {
	fg, bg := termbox.ColorDefault, termbox.ColorDefault

	termbox.Clear(fg, bg)
	ui.drawTitleLine()
	ui.drawMarketWin()
	ui.drawLabelWin()
	ui.drawStockWin()

	termbox.Flush()
}
func (ui *Ui) Prompt(cmd rune) {
	ui.lineEditor.Prompt(cmd, ui.selectedQuote)
}

func (ui *Ui) ExecuteCommand() {
	// execute some command
	oldQuoteId := ui.lineEditor.Execute(ui.selectedQuote)
	ui.lineEditor.Done()
	if ui.lineEditor.cmd == 'a' {
		ui.GetQuotes()
	}

	ui.resetSelection((*ui.stockQuotes)[oldQuoteId])
	ui.Draw()
}

func (ui *Ui) HandleInputKey(ev termbox.Event) {
	ui.lineEditor.Handle(ev)
}

func (ui *Ui) OpenInBrowser() {
	var err error
	q := (*ui.stockQuotes)[ui.selectedQuote]
	url := "https://wallmine.com/" + q.Ticker

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}
}

func (ui *Ui) NavigateLabel(key rune) {
	if key == 'h' || key == 'b' {
		ui.navigateLabelLeft()
	} else if key == 'l' || key == 'e' {
		ui.navigateLabelRight()
	} else if key == '0' {
		ui.selectedLabel = 0
	} else if key == '$' {
		ui.selectedLabel = len(ui.layout.columns) - 1
	}
	ui.Draw()
}

func (ui *Ui) SortLabel(key rune) {
	if key == 'j' {
		ui.sortSymbol = DESCENDING_CHAR
		ui.sortByLabelDsc()
	} else if key == 'k' {
		ui.sortSymbol = ASCENDING_CHAR
		ui.sortByLabelAsc()
	}
	ui.Draw()
}

func (ui *Ui) navigateLabelLeft() {
	if ui.selectedLabel > 0 {
		ui.selectedLabel -= 1
	}
}

func (ui *Ui) navigateLabelRight() {
	if ui.selectedLabel < len(ui.layout.columns)-1 {
		ui.selectedLabel += 1
	}
}

func (ui *Ui) sortByLabelDsc() {
	oldQ := (*ui.stockQuotes)[ui.selectedQuote]

	sort.SliceStable(*ui.stockQuotes, func(i, j int) bool {
		q := reflect.ValueOf((*ui.stockQuotes)[i]).Field(ui.selectedLabel).Interface()
		r := reflect.ValueOf((*ui.stockQuotes)[j]).Field(ui.selectedLabel).Interface()
		f1, isFloat := q.(float64)

		if isFloat {
			f2, _ := r.(float64)
			return f1 > f2
		} else {
			str, _ := q.(string)
			str2, _ := r.(string)
			return str > str2
		}
	})
	ui.resetSelection(oldQ)

}

func (ui *Ui) sortByLabelAsc() {
	oldQ := (*ui.stockQuotes)[ui.selectedQuote]
	sort.SliceStable(*ui.stockQuotes, func(i, j int) bool {
		q := reflect.ValueOf((*ui.stockQuotes)[i]).Field(ui.selectedLabel).Interface()
		r := reflect.ValueOf((*ui.stockQuotes)[j]).Field(ui.selectedLabel).Interface()
		f1, isFloat := q.(float64)

		if isFloat {
			f2, _ := r.(float64)
			return f1 < f2
		} else {
			str, _ := q.(string)
			str2, _ := r.(string)
			return str < str2
		}
	})
	ui.resetSelection(oldQ)

}

func (ui *Ui) resetSelection(oldQ Quote) {
	for id, _ := range *ui.stockQuotes {
		if (*ui.stockQuotes)[id] == oldQ {
			if id < ui.zerothQuote {
				// case if selected quote above current window
				for id < ui.zerothQuote {
					ui.navigateStockUp()
				}
			} else if id >= ui.zerothQuote+ui.stockWin.h {
				// case if selected quote is below current window
				// Do nothing.
				for id >= ui.zerothQuote+ui.stockWin.h {
					ui.navigateStockDown()
				}
			} else {
				// case if selected quote is exists in current window
				ui.selectedQuote = id
				ui.selectedVisibleQuote = ui.selectedQuote - ui.zerothQuote
			}
			break
		}
	}
}

// Temp for playing aroudn with termbox
func (ui *Ui) drawTitleLine() {
	fg, bg := termbox.ColorDefault|termbox.AttrBold, termbox.ColorDefault

	currentTime := time.Now()

	timeString := currentTime.Format(time.UnixDate)

	// %v and -%v for right and left justification respectively
	title := fmt.Sprintf("%-*v%*v", ui.titleWin.w/2, appTitle, ui.titleWin.w/2, timeString)

	ui.titleWin.print(0, 0, fg, bg, title)
}

func (ui *Ui) drawLabelWin() {
	fg, bg := termbox.ColorDefault|termbox.AttrUnderline, termbox.ColorDefault

	x := 0
	var label string

	for id, col := range ui.layout.columns {
		if id == ui.selectedLabel && *ui.mode == SORT {
			label = fmt.Sprintf("%-*v", col.width, col.name+" "+ui.sortSymbol)
			ui.labelWin.print(x, 0, termbox.ColorBlack, termbox.ColorWhite, label)
		} else if id == ui.selectedLabel && *ui.mode != SORT {
			// TODO: draw arrow based on sort typea (asc/dsc)
			label = fmt.Sprintf("%-*v", col.width, col.name+" "+ui.sortSymbol)
			ui.labelWin.print(x, 0, fg, bg, label)
		} else {
			label = fmt.Sprintf("%-*v", col.width, col.name)
			ui.labelWin.print(x, 0, fg, bg, label)
		}
		x += utf8.RuneCountInString(label)
	}

}

func (ui *Ui) drawMarketWin() {
	fg, bg := termbox.ColorDefault, termbox.ColorDefault

	x := 0
	y := 0
	for _, q := range *ui.marketQuotes {
		humanFormatted := float2Str(q.LastTrade, 2)
		tickerLine := fmt.Sprintf("%s ", marketNames[q.Ticker], humanFormatted, q.ChangePct)
		if x+len(tickerLine) > ui.marketWin.w {
			y++
			x = 0
		}
		indexLabel := fmt.Sprintf("%s ", marketNames[q.Ticker])
		ui.marketWin.print(x, y, termbox.ColorYellow, bg, indexLabel)
		x += len(indexLabel)
		changeLabel := fmt.Sprintf("%s (%.2f%%)  ", humanFormatted, q.ChangePct)
		ui.marketWin.print(x, y, fg, bg, changeLabel)
		x += len(changeLabel)

	}
}

func (ui *Ui) drawStockWin() {
	_, bg := termbox.ColorDefault, termbox.ColorDefault

	// TODO: don't do this. use a struct with properties, or some constants.
	for id, q := range ui.visibleQuotes {
		tickerLine := ""
		highlightColor := bg

		var lineColor termbox.Attribute
		if q.Change > 0 {
			lineColor = termbox.ColorGreen
		} else if q.Change == 0 {
			lineColor = termbox.ColorBlue
		} else {
			lineColor = termbox.ColorRed
		}

		if ui.selectedVisibleQuote == id && *ui.mode != SORT {
			highlightColor = termbox.ColorWhite
			lineColor = termbox.ColorBlack
		}

		v := reflect.ValueOf(q)

		for i := 0; i < v.NumField(); i++ {
			fieldVal := v.Field(i).Interface()
			val, ok := fieldVal.(float64)
			if ok {
				humanFormatted := float2Str(val, ui.layout.columns[i].precision)
				if strings.Contains(ui.layout.columns[i].name, "Dividend") {
					// get dividend yield %
					val = (val / q.LastTrade) * 100
				}
				if (strings.Contains(ui.layout.columns[i].name, "Change") || strings.Contains(ui.layout.columns[i].name, "After") || strings.Contains(ui.layout.columns[i].name, "Pre")) && val >= 0 {
					// TODO: just add an "advancing" field in Quote
					humanFormatted = "+" + humanFormatted
					tickerLine = tickerLine + fmt.Sprintf("%-*v", ui.layout.columns[i].width, humanFormatted)
				} else {
					tickerLine = tickerLine + fmt.Sprintf("%-*v", ui.layout.columns[i].width, humanFormatted)
				}
			} else {
				tickerLine = tickerLine + fmt.Sprintf("%-*v", ui.layout.columns[i].width, v.Field(i).Interface())
			}
		}

		ui.stockWin.print(0, id, lineColor, highlightColor, tickerLine)
	}
}

func (ui *Ui) GetQuotes() {
	var err error
	ui.stockQuotes, err = FetchQuotes(ui.profile.Tickers)
	if err != nil {
		fmt.Printf("error : %v", err)
		panic(err)
	}

	ui.marketQuotes, err = FetchMarket()

	if err != nil {
		panic(err)
	}
	// update stock window size to make life easier for us
	// bad practice? idc
	if len(*ui.stockQuotes) > ui.maxQuotesHeight {
		ui.stockWin.h = ui.maxQuotesHeight
		ui.visibleQuotes = (*ui.stockQuotes)[ui.zerothQuote : ui.zerothQuote+ui.maxQuotesHeight]
	} else {
		ui.stockWin.h = len(*ui.stockQuotes)
		// TODO: bug when deleting last quote?
		ui.visibleQuotes = *ui.stockQuotes
	}
	ui.lineEditor.quotes = ui.stockQuotes

	if ui.sortSymbol == DESCENDING_CHAR {
		ui.SortLabel('j')
	} else if ui.sortSymbol == ASCENDING_CHAR {
		ui.SortLabel('k')
	}
}

func (ui *Ui) navigateStockBeginning() {
	ui.zerothQuote = 0
	ui.selectedQuote = 0
	ui.selectedVisibleQuote = 0
	ui.visibleQuotes = (*ui.stockQuotes)[0:ui.stockWin.h]
}

func (ui *Ui) navigateStockEnd() {
	// surely this can be refactored/simplified
	ui.zerothQuote = len(*ui.stockQuotes) % ui.stockWin.h
	ui.selectedQuote = len(*ui.stockQuotes) - 1
	ui.selectedVisibleQuote = ui.stockWin.h - 1
	ui.visibleQuotes = (*ui.stockQuotes)[ui.zerothQuote:len(*ui.stockQuotes)]
	// fmt.Printf("zeroth: %d selected: %d visibleSelected: %d  visible: %v maxQuotesHeight: %d lenQuotes %d", ui.zerothQuote, ui.selectedQuote, ui.selectedVisibleQuote, ui.visibleQuotes, ui.maxQuotesHeight, len(*ui.stockQuotes))
}

func (ui *Ui) navigateStockDown() {
	// navigate down a line in the stock window
	updatedPos := ui.selectedQuote + 1
	if ui.selectedVisibleQuote+1 < ui.stockWin.h {
		ui.selectedQuote = updatedPos
		ui.selectedVisibleQuote += 1
	} else if ui.selectedVisibleQuote+1 >= ui.stockWin.h && updatedPos < len(*ui.stockQuotes) {
		ui.zerothQuote += 1
		ui.visibleQuotes = (*ui.stockQuotes)[updatedPos-ui.stockWin.h+1:]
		ui.selectedQuote = updatedPos
	}
}

func (ui *Ui) navigateStockUp() {
	// navigate up a line in the stock window
	updatedPos := ui.selectedQuote - 1
	if ui.selectedVisibleQuote-1 >= 0 {
		ui.selectedQuote = updatedPos
		ui.selectedVisibleQuote -= 1
	} else if updatedPos < ui.zerothQuote && ui.zerothQuote > 0 {
		ui.zerothQuote -= 1
		ui.selectedQuote = updatedPos
		ui.visibleQuotes = (*ui.stockQuotes)[ui.zerothQuote : ui.zerothQuote+ui.stockWin.h]
	}
}
