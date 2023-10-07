package lexer

import "fmt"

type Color int

const (
	// Colors
	COLOR_BLACK Color = iota
	COLOR_RED
	COLOR_GREEN
	COLOR_YELLOW
	COLOR_BLUE
	COLOR_MAGENTA
	COLOR_CYAN
	COLOR_WHITE
	COLOR_BRIGHT_BLACK
	COLOR_BRIGHT_RED
	COLOR_BRIGHT_GREEN
	COLOR_BRIGHT_YELLOW
	COLOR_BRIGHT_BLUE
	COLOR_BRIGHT_MAGENTA
	COLOR_BRIGHT_CYAN
	COLOR_BRIGHT_WHITE
)

var colorNames = map[Color]string{
	COLOR_BLACK:          "black",
	COLOR_RED:            "red",
	COLOR_GREEN:          "green",
	COLOR_YELLOW:         "yellow",
	COLOR_BLUE:           "blue",
	COLOR_MAGENTA:        "magenta",
	COLOR_CYAN:           "cyan",
	COLOR_WHITE:          "white",
	COLOR_BRIGHT_BLACK:   "bright_black",
	COLOR_BRIGHT_RED:     "bright_red",
	COLOR_BRIGHT_GREEN:   "bright_green",
	COLOR_BRIGHT_YELLOW:  "bright_yellow",
	COLOR_BRIGHT_BLUE:    "bright_blue",
	COLOR_BRIGHT_MAGENTA: "bright_magenta",
	COLOR_BRIGHT_CYAN:    "bright_cyan",
	COLOR_BRIGHT_WHITE:   "bright_white",
}

var colorValues = map[string]Color{
	"black":          COLOR_BLACK,
	"red":            COLOR_RED,
	"green":          COLOR_GREEN,
	"yellow":         COLOR_YELLOW,
	"blue":           COLOR_BLUE,
	"magenta":        COLOR_MAGENTA,
	"cyan":           COLOR_CYAN,
	"white":          COLOR_WHITE,
	"bright_black":   COLOR_BRIGHT_BLACK,
	"bright_red":     COLOR_BRIGHT_RED,
	"bright_green":   COLOR_BRIGHT_GREEN,
	"bright_yellow":  COLOR_BRIGHT_YELLOW,
	"bright_blue":    COLOR_BRIGHT_BLUE,
	"bright_magenta": COLOR_BRIGHT_MAGENTA,
	"bright_cyan":    COLOR_BRIGHT_CYAN,
	"bright_white":   COLOR_BRIGHT_WHITE,
}

func (c Color) String() string {
	if name, ok := colorNames[c]; ok {
		return name
	}
	return fmt.Sprintf("Color(%d)", c)
}
