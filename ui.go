package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"reflect"
	"runtime"
	"sort"
	"strconv"
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
	"BTC-USD":  "Bitcoin",
}

const (
	NO_CHAR         string = " "
	DESCENDING_CHAR string = "🠗"
	ASCENDING_CHAR  string = "🠕"
)
const (
	layoutUS = "01/02/2006"
)

const appTitle = "monmop 0.5"

const (
	titleWinHeight   int = 1
	marketWinHeight  int = 4
	labelWinHeight   int = 1
	commandWinHeight int = 1
)

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

	return &Ui{
		titleWin: &Win{
			w: wtot,
			h: titleWinHeight,
			x: 0,
			y: 0,
		},
		marketWin: &Win{
			w: wtot,
			h: marketWinHeight,
			x: 0,
			y: 1,
		},
		labelWin: &Win{
			w: wtot,
			h: labelWinHeight,
			x: 0,
			y: titleWinHeight + marketWinHeight,
		},
		stockWin: &Win{
			w: wtot,
			h: htot - (titleWinHeight + marketWinHeight + labelWinHeight + commandWinHeight),
			x: 0,
			y: 6,
		},
		commandWin: &Win{
			w: wtot,
			h: commandWinHeight,
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
	fg, bg := termbox.ColorDefault, termbox.ColorDefault
	termbox.Clear(fg, bg)
	wtot, htot := termbox.Size()
	ui.titleWin.w = wtot
	ui.marketWin.w = wtot
	ui.labelWin.w = wtot
	ui.stockWin.w = wtot
	ui.stockWin.h = htot - (ui.titleWin.h + ui.marketWin.h + ui.commandWin.h + ui.labelWin.h)
	ui.commandWin.w = wtot
	ui.commandWin.y = htot - 1
	ui.maxQuotesHeight = htot - (ui.titleWin.h + ui.marketWin.h + ui.commandWin.h + ui.labelWin.h)

	if ui.maxQuotesHeight < 0 {
		ui.maxQuotesHeight = 0
	}

	if len(ui.visibleQuotes) > ui.maxQuotesHeight {
		ui.visibleQuotes = ui.visibleQuotes[:ui.maxQuotesHeight]
	}

	ui.Clear()
	ui.Draw()
}

func (ui *Ui) Draw() {
	ui.drawTitleLine()
	ui.drawMarketWin()
	ui.drawLabelWin()
	ui.drawStockWin()

	termbox.Flush()
}

func (ui *Ui) Clear() {
	ui.titleWin.Clear()
	ui.stockWin.Clear()
	ui.labelWin.Clear()
	ui.commandWin.Clear()
}

func (ui *Ui) Prompt(cmd rune) {
	ui.lineEditor.Done() // clear the buffer
	ui.lineEditor.Prompt(cmd, ui.selectedQuote)
}

func (ui *Ui) ExecuteCommand() {
	switch ui.lineEditor.cmd {
	case 'a':
		tickerName := ui.lineEditor.AddQuotes()
		ui.GetQuotes()
		newQ := ui.getQuoteByTicker(tickerName)
		if newQ != nil {
			ui.updateSelection(*newQ)
		}
		ui.lineEditor.Done()
	case 'd':
		oldQuoteId := ui.lineEditor.Execute(ui.selectedQuote)
		ui.lineEditor.Done()
		ui.stockWin.Clear()
		ui.GetQuotes()
		if oldQuoteId >= 0 {
			ui.updateSelection((*ui.stockQuotes)[oldQuoteId])
		}
		ui.lineEditor.Done()
	case '/':
		oldQuoteId := ui.lineEditor.Execute(ui.selectedQuote)
		tickerName := ui.lineEditor.input
		ui.lineEditor.Done()
		if oldQuoteId < 0 {
			ui.lineEditor.Printf("couldn't find specified ticker(s): %s ", tickerName)
		} else {
			ui.updateSelection((*ui.stockQuotes)[oldQuoteId])
		}
	case ':':
		ui.lineEditor.Execute(ui.selectedQuote)
		ui.stockWin.Clear()
		ui.resetSelection()
		ui.GetQuotes()
	}
	ui.Draw()
}

func (ui *Ui) HandleLineEditorInput(ev termbox.Event) {
	ui.lineEditor.Handle(ev)
}

func (ui *Ui) OpenInBrowser() {
	var err error
	q := (*ui.stockQuotes)[ui.selectedQuote]
	url := "https://finance.yahoo.com/quote/" + q.Ticker

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

func (ui *Ui) HandleSortEvent(key rune) {
	if key == 'h' || key == 'b' {
		ui.navigateLabelLeft()
	} else if key == 'l' || key == 'e' {
		ui.navigateLabelRight()
	} else if key == '0' {
		ui.selectedLabel = 0
	} else if key == '$' {
		ui.selectedLabel = len(ui.layout.columns) - 1
	}

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
	if len(*ui.stockQuotes) == 0 {
		return
	}

	oldQ := (*ui.stockQuotes)[ui.selectedQuote]

	sort.SliceStable(*ui.stockQuotes, func(i, j int) bool {
		q := reflect.ValueOf((*ui.stockQuotes)[i]).Field(ui.selectedLabel).Interface()
		r := reflect.ValueOf((*ui.stockQuotes)[j]).Field(ui.selectedLabel).Interface()
		f1, isFloat := q.(float64)
		j1, isJsonNumber := q.(json.Number)

		if isFloat {
			f2, _ := r.(float64)
			return f1 > f2
		} else if isJsonNumber {
			j2, _ := r.(json.Number)
			return j1 > j2
		} else {
			str, _ := q.(string)
			str2, _ := r.(string)
			return str > str2
		}
	})
	ui.profile.Tickers = ui.getSortedTickers(*ui.stockQuotes)
	ui.updateSelection(oldQ)

}

func (ui *Ui) sortByLabelAsc() {
	if len(*ui.stockQuotes) == 0 {
		return
	}

	oldQ := (*ui.stockQuotes)[ui.selectedQuote]
	sort.SliceStable(*ui.stockQuotes, func(i, j int) bool {
		q := reflect.ValueOf((*ui.stockQuotes)[i]).Field(ui.selectedLabel).Interface()
		r := reflect.ValueOf((*ui.stockQuotes)[j]).Field(ui.selectedLabel).Interface()
		f1, isFloat := q.(float64)
		j1, isJsonNumber := q.(json.Number)

		if isFloat {
			f2, _ := r.(float64)
			return f1 < f2
		} else if isJsonNumber {
			j2, _ := r.(json.Number)
			return j1 < j2
		} else {
			str, _ := q.(string)
			str2, _ := r.(string)
			return str < str2
		}
	})
	ui.profile.Tickers = ui.getSortedTickers(*ui.stockQuotes)
	ui.updateSelection(oldQ)
}

func (ui *Ui) updateSelection(newQ Quote) {
	for id := range *ui.stockQuotes {
		if (*ui.stockQuotes)[id] == newQ {
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

func (ui *Ui) resetSelection() {
	// must be called in conjunction with GetQuotes()
	if len(*ui.stockQuotes) == 0 {
		return
	}
	ui.zerothQuote = 0
	ui.selectedVisibleQuote = 0
	ui.selectedQuote = 0
}

func (ui *Ui) getSortedTickers(quotes []Quote) []string {
	tickers := make([]string, len(quotes))
	for id, q := range quotes {
		tickers[id] = q.Ticker
	}
	return tickers
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
		tickerLine := fmt.Sprintf("%s %s %.2f", marketNames[q.Ticker], humanFormatted, q.ChangePct)
		if x+len(tickerLine) > ui.marketWin.w {
			y++
			x = 0
		}
		if y >= ui.marketWin.h-1 {
			break
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

	for id, q := range ui.visibleQuotes {
		tickerLine := ""
		highlightColor := bg
		if q.Ticker == "" {
			// TODO BAD antipattern, should fix this
			// im only doing this because "deleting" a quote
			// messes up the visible quotes slice in the edge case
			// where we are looking at the tail end of quotes
			// it works though...
			continue
		}

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
				if strings.Contains(ui.layout.columns[i].name, "Div") {
					// get dividend yield %
					val = val * 100
				}
				humanFormatted := float2Str(val, ui.layout.columns[i].precision)
				if (strings.Contains(ui.layout.columns[i].name, "Change") ||
					strings.Contains(ui.layout.columns[i].name, "After") ||
					strings.Contains(ui.layout.columns[i].name, "Pre")) && val >= 0 {
					// TODO: just add an "advancing" field in Quote
					humanFormatted = "+" + humanFormatted
					tickerLine = tickerLine + fmt.Sprintf("%-*v", ui.layout.columns[i].width, humanFormatted)
				} else {
					tickerLine = tickerLine + fmt.Sprintf("%-*v", ui.layout.columns[i].width, humanFormatted)
				}
			} else if strings.Contains(ui.layout.columns[i].name, "Earnings") {
				earningsTs := string(fieldVal.(json.Number))
				earningsTsInt, err := strconv.ParseInt(earningsTs, 10, 64)
				earningsStr := "-"

				if err == nil {
					tm := time.Unix(earningsTsInt, 0)
					earningsStr = tm.Format(layoutUS)
				}

				tickerLine = tickerLine + fmt.Sprintf("%-*v", ui.layout.columns[i].width, earningsStr)

			} else {
				tickerLine = tickerLine + fmt.Sprintf("%-*v", ui.layout.columns[i].width, fieldVal)
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
	if len(*ui.stockQuotes) > ui.maxQuotesHeight {
		ui.stockWin.h = ui.maxQuotesHeight
	} else {
		ui.stockWin.h = len(*ui.stockQuotes)
	}

	ui.visibleQuotes = (*ui.stockQuotes)[ui.zerothQuote : ui.zerothQuote+ui.stockWin.h]
	ui.lineEditor.quotes = ui.stockQuotes

	if ui.sortSymbol == DESCENDING_CHAR {
		ui.HandleSortEvent('j')
	} else if ui.sortSymbol == ASCENDING_CHAR {
		ui.HandleSortEvent('k')
	}
}

func (ui *Ui) navigateStockBeginning() {
	for ui.selectedQuote > 0 {
		ui.navigateStockUp()
	}
}

func (ui *Ui) navigateStockEnd() {
	for ui.selectedQuote < len(*ui.stockQuotes)-1 {
		ui.navigateStockDown()
	}
}

func (ui *Ui) navigateStockDown() {
	updatedPos := ui.selectedQuote + 1
	if updatedPos >= len(*ui.stockQuotes) || len(*ui.stockQuotes) == 0 {
		return
	}

	if ui.selectedVisibleQuote+1 < ui.stockWin.h {
		ui.selectedQuote = updatedPos
		ui.selectedVisibleQuote += 1
	} else if ui.selectedVisibleQuote+1 >= ui.stockWin.h {
		ui.zerothQuote += 1
		ui.visibleQuotes = (*ui.stockQuotes)[ui.zerothQuote : ui.zerothQuote+ui.stockWin.h]
		ui.selectedQuote = updatedPos
	}
}

func (ui *Ui) navigateStockUp() {
	updatedPos := ui.selectedQuote - 1
	if updatedPos < 0 || len(*ui.stockQuotes) == 0 {
		return
	}

	if ui.selectedVisibleQuote-1 >= 0 {
		ui.selectedQuote = updatedPos
		ui.selectedVisibleQuote -= 1
	} else if updatedPos < ui.zerothQuote && ui.zerothQuote > 0 {
		ui.zerothQuote -= 1
		ui.selectedQuote = updatedPos
		ui.visibleQuotes = (*ui.stockQuotes)[ui.zerothQuote : ui.zerothQuote+ui.stockWin.h]
	}
}

func (ui *Ui) getQuoteByTicker(ticker string) *Quote {
	for id, q := range *ui.stockQuotes {
		if q.Ticker == ticker {
			return &(*ui.stockQuotes)[id]
		}
	}
	return nil
}
