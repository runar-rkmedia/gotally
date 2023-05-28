package dev

import (
	"fmt"
	"path"
	"runtime"

	"github.com/gookit/color"
)

func Printf(format string, a ...any) {
	_, file, line, _ := runtime.Caller(1)
	color.Printf("<comment>%s:%d</> %s", path.Base(file), line, fmt.Sprintf(format, a...))
}
func Println(a ...any) {
	_, file, line, _ := runtime.Caller(1)
	color.Printf("<comment>%s:%d</> <lightCyan>%s\n</>", path.Base(file), line, fmt.Sprint(a...))
}
