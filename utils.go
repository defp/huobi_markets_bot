package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"strings"
)

func times(str string, n int) (out string) {
	for i := 0; i < n; i++ {
		out += str
	}
	return
}

// Right right-pads the string with pad up to len runes
func padRight(str string, length int, pad string) string {
	return str + times(pad, length-len(str))
}

func formatChange(change float64) string {
	var riseText string
	if change > 0 {
		riseText = fmt.Sprintf("+%.2f%%", change)
	} else if change < 0 {
		riseText = fmt.Sprintf("%.2f%%", change)
	} else {
		riseText = fmt.Sprintf("=%.2f%%", change)
	}
	return riseText
}

func unzip(data []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(r)
}

func getCoinText(mainCoin, coin, format string) string {
	td := coinClosePrice[coin+mainCoin]
	coinName := padRight(strings.ToUpper(coin), 4, " ")
	change := (td.Close - td.Open) / td.Open * 100
	closePrice := padRight(fmt.Sprintf(format, td.Close), 10, " ")

	riseText := formatChange(change)
	return fmt.Sprintf("%s $%s %s\n", coinName, closePrice, riseText)
}
