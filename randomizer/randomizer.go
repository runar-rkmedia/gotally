package randomizer

import (
	"hash/maphash"
	"sync"

	"github.com/runar-rkmedia/gotally/fastrand"
)

type randomizer struct {
	ex          *fastrand.PCG64
	initialSeed uint64
	name        string
	sync.RWMutex
}

func NewSeededRandomizer() *randomizer {

	return NewRandomizerFromSeed(new(maphash.Hash).Sum64(), new(maphash.Hash).Sum64())
}

// Thread-safe randomizer.
// Ideally it should be fast and seedable, but should also work identical across platforms
// and programming languages. Currently, this implementation is only fast and thread-safe.
func NewRandomizer(seed uint64) *randomizer {
	return NewRandomizerFromSeed(seed, 0)
}
func NewRandomizerFromSeed(seed, state uint64) *randomizer {
	src := fastrand.NewPCG64(seed, state)
	r := randomizer{src, seed, "", sync.RWMutex{}}
	return &r
}

func (r *randomizer) SetName(s string) {
	r.Lock()
	defer r.Unlock()
	r.name = s

}
func (r *randomizer) Intn(n int) int {
	return r.Int() % n
}
func (r *randomizer) Int63n(n int64) int64 {
	return r.Int63() % n
}
func (r *randomizer) Uint63() uint64 {
	r.Lock()
	defer r.Unlock()
	return r.ex.Uint64()
}
func (r *randomizer) Seed() (uint64, uint64) {
	r.Lock()
	defer r.Unlock()
	return r.ex.GetSeed()
	// return r.src.State
}
func (r *randomizer) SetSeed(a, b uint64) error {
	r.Lock()
	defer r.Unlock()
	r.initialSeed = a
	r.ex.Seed(a, b)
	return nil
}
func (r *randomizer) Int63() int64 {
	u := int64(r.Uint63())
	if u < 0 {
		u *= -1
	}
	return u
}
func (r *randomizer) Int() int {
	return int(r.Int63())
}
