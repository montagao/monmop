package main

// formatting rules for indiviudal column within a stock/lable win

type Column struct {
	width     int
	name      string
	precision int
}

type Layout struct {
	columns []Column
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
