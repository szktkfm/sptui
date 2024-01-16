package sptui

import (
	"fmt"
	"strings"

	"github.com/mattn/go-runewidth"
	"golang.org/x/text/unicode/norm"
)

const (
	// ZWSP represents zero-width space.
	ZWSP = '\u200B'

	// ZWNBSP represents zero-width no-break space.
	ZWNBSP = '\uFEFF'

	// ZWJ represents zero-width joiner.
	ZWJ = '\u200D'

	// ZWNJ represents zero-width non-joiner.
	ZWNJ = '\u200C'

	empty = ""
)

var replacer = strings.NewReplacer(string(ZWSP), empty)

func RemoveZeroWidthSpace(s string) string {
	return strings.Replace(s, string(ZWSP), empty, -1)
}

// Normalize form C
func Normalize(s string) string {
	return norm.NFC.String(s)
}

func CleanString(s string) string {
	s = RemoveZeroWidthSpace(s)
	s = Normalize(s)
	return s
}

func WrapText(s string, width int, line int) string {
	var ret string
	cnt := 0
	length := runewidth.StringWidth(s)
	for {
		if cnt == line-1 {
			ret += runewidth.Truncate(s, width, "...")
			break
		}
		ret += runewidth.Truncate(s, width, "")
		s = runewidth.TruncateLeft(s, width, "")
		cnt++
		if length-cnt*listWidth <= 0 {
			break
		}
		ret += "\n"
	}
	return ret
}

func PadOrTruncate(s string, n int) string {
	if runewidth.StringWidth(s) > listWidth {
		return fmt.Sprintf("%s", runewidth.Truncate(s, n, ""))
	} else {
		return s + strings.Repeat(" ", listWidth-runewidth.StringWidth(s))
	}
}
