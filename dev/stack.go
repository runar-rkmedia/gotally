package dev

import (
	"regexp"
	"runtime/debug"
	"strings"
)

var nameReg = regexp.MustCompile(`[^ ]*\/gotally\/`)
var addressReg = regexp.MustCompile(`[\+_ ,\{]*0x[^ ]*`)
var cleanUpReg = regexp.MustCompile(`(\( |\(?_, ?|\{*_, ?| ?_},?|github\.com|\.{3}|\}?\))`)

func Stack() string {
	stackLines := strings.Split(string(debug.Stack()), "\n")
	filtered := []string{}
	paddingTarget := 30
	for i := 5; i < len(stackLines); i += 2 {
		if strings.Contains(stackLines[i], "gotally") {
			f1 := cleanUpReg.ReplaceAllString(
				addressReg.ReplaceAllString(
					nameReg.ReplaceAllString(stackLines[i+1],
						""),
					""),
				"")
			f2 := cleanUpReg.ReplaceAllString(
				addressReg.ReplaceAllString(
					nameReg.ReplaceAllString(stackLines[i],
						""),
					""),
				"")
			paddingLength := paddingTarget - len(f1)
			padding := ""
			for i := 0; i < paddingLength; i++ {
				padding += " "

			}

			filtered = append(filtered, f1+padding+f2)
		}

	}

	return strings.Join(filtered, "\n") + "\n"
}
