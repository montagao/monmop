package yahoo

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQuotes(t *testing.T) {
	tickers := []string{"GOOG", "BA"}
	quotes, err := FetchAll(tickers)
	assert.NoError(t, err)
	fmt.Printf("quotes %v", quotes)
}
