package tallylogic

import (
	"testing"
)

func Benchmark_Randomizer(b *testing.B) {
	r := NewRandomizer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			r.Intn(1000)
		}
	})
}
