package main

import (
	"bytes"
	"fmt"
	"log"
	"os"

	"github.com/runar-rkmedia/gotally/namegenerator"
	flag "github.com/spf13/pflag"
)

func main() {
	count := flag.IntP("count", "c", 1, "count of names to generate")
	min := flag.Int("min-length", 0, "minimum length")
	max := flag.Int("max-length", 0, "maximum length")
	length := flag.Int("length", 0, "exact length")
	randomizer := flag.String("randomizer", "random", "which randomizer to use. One of: random, consecutive")
	out := flag.StringP("out", "o", "", "Write to file instead of stdout")
	lineSep := flag.String("name-seperator", "\n", "Seperator to use after names.")
	spaceSep := flag.String("space-seperator", " ", "Seperator to use between words in the name.")
	printStats := flag.Bool("print-stats", false, "print stats")
	printEntropy := flag.Bool("print-entropy", false, "print combined entropy of wordlists")
	all := flag.Bool("all", false, "sets count to be equal to the combined entropy")
	seed := flag.Uint64("seed", 0, "Set seed")
	flag.Usage = func() {
		fmt.Printf("namegenerator generates randomized names based on dictionaries\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if length != nil && *length != 0 {
		min = length
		max = length
	}
	var gen namegenerator.NameGenerator
	switch *randomizer {
	case "random":
		gen = namegenerator.NewNameGenerator()
	case "consecutive":
		gen = namegenerator.NewNameGeneratorCensucutive()
	default:
		log.Fatal("randomizer must be one of: random, consecutive")
	}
	if seed != nil && *seed != 0 {
		gen.SetSeed(*seed, 0)
	}
	if spaceSep != nil {
		gen.SetSeparator(*spaceSep)
	}
	if *all {
		*count = gen.CombinedEntropy()
	}
	var buf bytes.Buffer
	for A := 0; A < *count; A++ {
		buf.WriteString(gen.NameAtLength(*min, *max) + *lineSep)
	}
	if *out != "" {
		os.WriteFile(*out, buf.Bytes(), 0644)
	} else {
		os.Stdout.Write(buf.Bytes())
	}

	if *printEntropy {
		fmt.Println(gen.CombinedEntropy())
	}
	if *printStats {
		fmt.Printf("%#v", gen.Stats())
	}
}
