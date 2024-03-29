## monmop

monmop is inspired by [mop](https://github.com/mop-tracker/mop), which is a command line utility for tracking the stock market.

![Screenshot](https://raw.githubusercontent.com/montagao/monmop/master/docs/screenshot.png "monmop")

This fork aims to add vim-like navigation and editing and being able to open up a particular stock on yahoo finance for more details.

### Installing monmop
requires go v1.13+
```
# Make sure your $GOPATH is set and $GOPATH/bin is in your $PATH
$ go get github.com/montagao/monmop
$ monmop
```

### Using monmop
Keyboard Shortcuts:
```
o/Enter - open detailed page about selected ticker in browser
j/k - navigate up or down
a - add a list of comma separated tickers
d - delete currently selected ticker
/ - search for a ticker
q - quit monmop
s - sort stock by label
```

### Configuration:

By default the list of tickers is saved/read from `~/.config/monmop/monmoprc`
