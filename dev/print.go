package dev

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path"
	"runtime"

	"github.com/ghodss/yaml"
	"github.com/gookit/color"
	"github.com/pelletier/go-toml/v2"
)

func Printf(format string, a ...any) {
	_, file, line, _ := runtime.Caller(1)
	color.Printf("<comment>%s:%d</> %s", path.Base(file), line, fmt.Sprintf(format, a...))
}
func Println(a ...any) {
	_, file, line, _ := runtime.Caller(1)
	color.Printf("<comment>%s:%d</> <lightCyan>%s</>", path.Base(file), line, fmt.Sprintln(a...))
}

func ToJson(a any) string {
	b, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(b)
}
func ToYaml(a any) string {
	b, err := yaml.Marshal(a)
	if err != nil {
		panic(err)
	}
	return string(b)
}
func ToToml(a any) string {

	buf := bytes.Buffer{}
	enc := toml.NewEncoder(&buf)
	err := enc.Encode(a)
	if err != nil {
		panic(err)
	}
	return buf.String()
}

func printValue(skip int, name string, value string) {
	_, file, line, _ := runtime.Caller(skip)
	color.Printf("<comment>%s:%d</> <lightCyan>%s: %s\n</>", path.Base(file), line, name, value)
}
func PrintJson(name string, value any) {
	printValue(2, name, ToJson(value))
}
func PrintYaml(name string, value any) {
	printValue(2, name, ToYaml(value))
}
func PrintToml(name string, value any) {
	printValue(2, name, ToToml(value))
}
