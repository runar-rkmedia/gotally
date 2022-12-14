package randomizer

import (
	"testing"

	"github.com/go-test/deep"
)

func Benchmark_Randomizer(b *testing.B) {
	r := NewRandomizer(1)
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			r.Intn(1000)
		}
	})
}

func Test_randomizer_Intn(t *testing.T) {
	tests := []struct {
		name  string
		seed  uint64
		state uint64
		arg   int
		want  []int
	}{
		{
			"Simple test",
			123,
			123,
			144,
			[]int{120, 104, 95},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRandomizerFromSeed(tt.seed, tt.seed)
			got := make([]int, len(tt.want))
			for i := range tt.want {
				got[i] = r.Intn(tt.arg)
			}
			if diff := deep.Equal(got, tt.want); diff != nil {
				t.Errorf("Did not match: %v != %v", got, tt.want)
			}

		})
	}
}
