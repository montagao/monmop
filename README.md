## monmop

monmop is my own fork of [mop](https://github.com/mop-tracker/mop), which is a command line utility for tracking the stock market.

![Screenshot](https://raw.githubusercontent.com/montagao/monmop/master/docs/screenshot.png "monmop")

This fork aims to add vim-like navigation and editing and being able to expand a stock to further look at it's graphs (WIP) and financials.

### Installing monmop
requires go v1.13+
```
# Make sure your $GOPATH is set and $GOPATH/bin is in your $PATH
$ go get github.com/montagao/monmop
$ monmop
```

### Using monmop
```
o/Enter - open detailed page about selected ticker in browser
a - add a list of comma separated tickers
d - delete currently selected ticker
/ - search for a ticker
```
