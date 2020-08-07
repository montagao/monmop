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
		{9, `Ticker`, 0},
		{10, `Last`, 2},
		{10, `Change`, 2},
		{10, `Change %`, 2},
		{10, `Open`, 2},
		{10, `Low`, 2},
		{10, `High`, 2},
		{10, `Volume`, 2},
		{12, `Avg Volume`, 2},
		{10, `P/E`, 2},
		{9, `Divd %`, 2},
		{10, `Mkt Cap`, 3},
		{12, `Earnings`, 0},
		{11, `PreChg %`, 2},
		{11, `AfterChg %`, 2},
	}

	return layout
}
