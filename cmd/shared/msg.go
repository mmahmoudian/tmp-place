package shared

import (
	"fmt"
	"strings"
)

// Msg prints a formatted message to the console.
func Msg(color string, message string, args ...any) {
	const reset = "\033[0m"
	var colors = map[string]string{
		// very nice explanation here: https://stackoverflow.com/a/33206814/1613005
		"error":   "\033[41;97m ⛌ \033[0;31m ", // red
		"success": "\033[42;97m ✓ \033[0;32m ", // green
		"warning": "\033[43;97m ⚠ \033[0;33m ", // yellow
		"info":    "\033[44;97m ℹ \033[0;34m ", // blue

		"red":     "\033[31m",
		"green":   "\033[32m",
		"yellow":  "\033[33m",
		"blue":    "\033[34m",
		"magenta": "\033[35m",
		"cyan":    "\033[36m",
		"gray":    "\033[37m",
		"white":   "\033[97m",
	}

	if code, ok := colors[strings.ToLower(color)]; ok {
		// fmt.Println(code + fmt.Sprint(message, args) + reset)
		fmt.Println(code + message + fmt.Sprint(args...) + reset)
		return
	}
	fmt.Println(message, fmt.Sprint(args...))
}

// Msgf prints a formatted message with arguments to the console.
func Msgf(color string, message string, args ...any) {
	const reset = "\033[0m"
	var colors = map[string]string{
		"error":   "\033[31m", // red
		"success": "\033[32m", // green
		"warning": "\033[33m", // yellow
		"output":  "\033[34m", // blue

		"magenta": "\033[35m", // magenta
		"cyan":    "\033[36m", // cyan
		"gray":    "\033[37m", // gray
		"white":   "\033[97m", // white
	}

	if code, ok := colors[strings.ToLower(color)]; ok {
		fmt.Println(code + fmt.Sprintf(message, args...) + reset)
		return
	}
	fmt.Println(fmt.Sprintf(message, args...))
}
