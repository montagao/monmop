package main

// formatting rules for indiviudal column within a stock/lable win

type Column struct {
	width     int
	name      string
	precision int
}

type Layout struct {
	columns []Column

	//TODO:
	// sorter         *Sorter            // Pointer to sorting receiver.
	// filter         *Filter            // Pointer to filtering receiver.
	// regex          *regexp.Regexp     // Pointer to regular expression to align decimal points.
	// marketTemplate *template.Template // Pointer to template to format market data.
	// quotesTemplate *template.Template // Pointer to template to format the list of stock quotes.
}

func NewLayout() *Layout {
	layout := &Layout{}
	layout.columns = []Column{
		{10, `Ticker`, 0},
		{10, `Last`, 4},
		{10, `Change`, 4},
		{10, `Change %`, 2},
		{10, `Open`, 4},
		{10, `Low`, 4},
		{10, `High`, 4},
		{10, `Volume`, 2},
		{15, `Avg Volume`, 2},
		{10, `P/E`, 2},
		{15, `Dividend %`, 2},
		{15, `Mkt Cap`, 4},
		{15, `PreChg %`, 2},
		{15, `AfterChg %`, 2},
	}

	return layout
}
