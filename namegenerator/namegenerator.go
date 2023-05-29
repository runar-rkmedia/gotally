package namegenerator

import (
	"fmt"
	"math"
	"strings"
	"sync/atomic"

	"github.com/gookit/color"
	"github.com/runar-rkmedia/gotally/randomizer"
)

type randomNumberGenerator interface {
	Int() int
	SetSeed(a, b uint64) error
}

type NameGenerator struct {
	randomizer randomNumberGenerator
	seperator  string
}

type rn struct {
	c uint64
}

func (r *rn) Int() int {
	c := r.c
	// r.c++
	atomic.AddUint64(&r.c, 1)
	return int(c)
}
func (r *rn) SetSeed(a, b uint64) error {
	r.c = a
	return nil
}

func NewNameGenerator() NameGenerator {
	return NameGenerator{
		randomizer: randomizer.NewRandomizer(0),
		seperator:  " ",
	}
}
func NewNameGeneratorCensucutive() NameGenerator {
	if false {

		superlatives = superlatives[0:4]
		adjectives = adjectives[0:3]
		subjectives = subjectives[0:2]
		i := 0
		color.Printf("<notice>%02d | %d %d %d % 14s % 10s % 10s\n</>", i, i, i, i, "Superlative", "Adjective", "Subject")
		for a, sup := range superlatives {
			for b, adj := range adjectives {

				for c, sub := range subjectives {
					fmt.Printf("%02d | %d %d %d % 14s %10s % 10s\n", i, a, b, c, sup, adj, sub)
					i++

				}
			}

		}
	}
	return NameGenerator{
		randomizer: &rn{0},
	}
}

type stats struct {
	DictionaryCount map[string]int
	CombinedEntropy int
}

func (g NameGenerator) Stats() stats {
	return stats{
		CombinedEntropy: g.CombinedEntropy(),
		DictionaryCount: map[string]int{
			"adjectives":   len(adjectives),
			"superlatives": len(superlatives),
			"subjectives":  len(subjectives),
		},
	}
}
func (g *NameGenerator) Name() string {
	return g.SeededName(g.randomizer.Int())
}
func (g *NameGenerator) SeededName(seed int) string {
	s := g.name(seed)
	return s.String()
}
func (g *NameGenerator) SetSeed(seed, state uint64) error {
	return g.randomizer.SetSeed(seed, state)
}
func (g *NameGenerator) NameAtLength(min, max int) string {
	return g.nameAtLength(min, max, 100)
}
func floorDiv(a, b int) int {
	return int(math.Floor(float64(a) / float64(b)))
}
func (g *NameGenerator) name(seed int) strings.Builder {
	s := strings.Builder{}
	iSuperlative := (seed / (len(adjectives) * len(subjectives))) % len(superlatives)
	iAdjective := (seed / len(subjectives)) % len(adjectives)
	ISubjective := seed % len(subjectives)
	s.WriteString(superlatives[iSuperlative])
	s.WriteString(g.seperator)
	s.WriteString(adjectives[iAdjective])
	s.WriteString(g.seperator)
	s.WriteString(subjectives[ISubjective])
	s.WriteString(g.seperator)
	return s
}
func (g *NameGenerator) nameAtLength(min, max int, maxAttemepts int) string {
	seed := g.randomizer.Int()
	s := g.name(seed)

	if min == 0 && max == 0 {
		return s.String()
	}
	if maxAttemepts <= 0 {
		return s.String()
	}
	if min > 0 && s.Len() < min {
		maxAttemepts--
		return g.nameAtLength(min, max, maxAttemepts)
	}
	if max > 0 && max >= min && s.Len() > max {
		maxAttemepts--
		return g.nameAtLength(min, max, maxAttemepts)
	}

	return s.String()
}

func (g *NameGenerator) SetSeparator(sep string) {
	g.seperator = sep
}
func (g *NameGenerator) CombinedEntropy() int {
	return len(superlatives) * len(adjectives) * len(subjectives)
}
