package main

import (
	"fmt"
	"reflect"
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
	titleWin      *Win
	marketWin     *Win
	labelWin      *Win
	stockWin      *Win
	layout        *Layout
	selectedQuote int
	selectedSort  int
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
			h: 1,
			x: 0,
			y: 6,
		},
		layout:        NewLayout(),
		selectedQuote: 0,
		selectedSort:  0,
	}

}

func (ui *Ui) draw(marketQuotes *[]Quote, stockQuotes *[]Quote) {
	fg, bg := termbox.ColorDefault, termbox.ColorDefault

	termbox.Clear(fg, bg)
	ui.drawTitleLine()
	ui.drawMarketWin(marketQuotes)
	ui.drawLabelWin()
	ui.drawStockWin(stockQuotes)

	// w, h := termbox.Size()
	// fmt.Printf("termbox size: %d, %d", w, h)
	termbox.Flush()
}

// Temp for playing aroudn with termbox
func (ui *Ui) drawTitleLine() {
	fg, bg := termbox.ColorDefault|termbox.AttrBold|termbox.AttrUnderline, termbox.ColorDefault

	title := "Monmop by Monta"

	ui.titleWin.print(0, 0, fg, bg, title)
}

func (ui *Ui) drawLabelWin() {
	fg, bg := termbox.ColorDefault|termbox.AttrUnderline, termbox.ColorDefault

	labels := ""
	for _, col := range ui.layout.columns {
		labels = labels + fmt.Sprintf("%-*v", col.width, col.name)
	}

	ui.labelWin.print(0, 0, fg, bg, labels)
}

func (ui *Ui) drawMarketWin(quotes *[]Quote) {
	fg, bg := termbox.ColorDefault, termbox.ColorDefault

	x := 0
	y := 0
	for _, q := range *quotes {
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

func (ui *Ui) drawStockWin(quotes *[]Quote) {
	_, bg := termbox.ColorDefault, termbox.ColorDefault

	// TODO: don't do this. use a struct with properties, or some constants.
	for id, q := range *quotes {
		tickerLine := ""
		highlightColor := bg
		if ui.selectedQuote == id {
			highlightColor = termbox.ColorWhite
		}

		var lineColor termbox.Attribute
		if q.Change > 0 {
			lineColor = termbox.ColorGreen
		} else if q.Change == 0 {
			lineColor = termbox.ColorWhite
		} else {
			lineColor = termbox.ColorRed
		}

		v := reflect.ValueOf(q)

		for i := 0; i < v.NumField(); i++ {
			fieldVal := v.Field(i).Interface()
			val, ok := fieldVal.(float64)
			if ok {
				humanFormatted := float2Str(val, ui.layout.columns[i].precision)
				tickerLine = tickerLine + fmt.Sprintf("%-*v", ui.layout.columns[i].width, humanFormatted)
			} else {
				tickerLine = tickerLine + fmt.Sprintf("%-*v", ui.layout.columns[i].width, v.Field(i).Interface())
			}
		}

		ui.stockWin.print(0, id, lineColor, highlightColor, tickerLine)
	}
}

func (ui *Ui) navigateStock(jump int) {
	//TODO: think aobut how to write this properly

	// // navigate up or down a line in the stock window
	// if jump > 0 && ui.selectedQuote < len(*stockWin.height)-jump {
	// 	ui.selectedQuote += jump // make this a method?
	// } else if jump < 0 && (ui.selectedQuote+jump) >= 0 {
	// 	ui.selectedQuote += jump // make this a method?
	// }

}
