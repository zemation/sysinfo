package system

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"golang.org/x/term"
)

const (
	colorReset  = "\033[0m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
)

func IsTerminal() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

func Colorize(value string, warn, danger float64) string {
	if !IsTerminal() {
		return value
	}
	clean := strings.TrimSuffix(value, "%")
	num, err := strconv.ParseFloat(clean, 64)
	if err != nil {
		return value
	}
	switch {
	case num >= danger:
		return fmt.Sprintf("%s%s%s", colorRed, value, colorReset)
	case num >= warn:
		return fmt.Sprintf("%s%s%s", colorYellow, value, colorReset)
	default:
		return fmt.Sprintf("%s%s%s", colorGreen, value, colorReset)
	}
}