package main

import (
	"flag"
	"fmt"
	"strconv"

	"github.com/runar-rkmedia/gotally/tallylogic"
)

func main() {
	args := flag.Args()
	flag.Parse()
	fmt.Println(args)
	a, err := strconv.ParseInt(flag.Arg(0), 10, 64)
	if err != nil {
		panic(err)
	}
	b, err := strconv.ParseInt(flag.Arg(1), 10, 64)
	if err != nil {
		panic(err)
	}
	cell := tallylogic.NewCell(a, int(b))
	fmt.Println(cell)
}
