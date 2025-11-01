package shared

import (
	"fmt"
)

// AskString is a function to prompt the user for input in the CLI.
func AskString(prompt string) string {
	var response string
	fmt.Print(prompt)
	fmt.Scanln(&response)
	return response
}

// AskYesNo prompts the user for a yes/no confirmation in the CLI.
func AskYesNo(prompt string) bool {
	var response string
	for {
		fmt.Print(prompt + " (y/n): ")
		fmt.Scanln(&response)
		switch response {
		case "y", "Y":
			return true
		case "n", "N":
			return false
		default:
			fmt.Println("Please enter 'y' or 'n'. Ctrl-c to quit.")
		}
	}
}
