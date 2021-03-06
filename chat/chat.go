// Package chat contains the symbols for message formatting for convenience.
package chat

import (
	"strings"
)

// Constants for message formatting for convenience.
const (
	S = "§"

	Black      = S + "0"
	Blue       = S + "1"
	Green      = S + "2"
	Aqua       = S + "3"
	Red        = S + "4"
	Purple     = S + "5"
	Gold       = S + "6"
	LightGray  = S + "7"
	Gray       = S + "8"
	LightBlue  = S + "9"
	LightGreen = S + "a"
	LightAqua  = S + "b"
	LightRed   = S + "c"
	Pink       = S + "d"
	Yellow     = S + "e"
	White      = S + "f"

	Scramble      = S + "k"
	Bold          = S + "l"
	Strikethrough = S + "m"
	Underline     = S + "n"
	Italic        = S + "o"
	Reset         = S + "r"
)

// Format formats the message given by replacing all & with § so you can use
// color codes like "&4" for red.
//
// For example:
//
//     Format("&4Hello!")
//
// returns "§4Hello!"
func Format(message string) string {
	return strings.Replace(message, "&", S, -1)
}
