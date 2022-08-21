package tallylogic

import (
	"hash/maphash"
)

type randomizer struct{}

// Thread-safe randomizer.
// Ideally it should be fast and seedable, but should also work identical across platforms
// and programming languages. Currently, this implementation is only fast and thread-safe.
func NewRandomizer() randomizer {
	return randomizer{}
}

func (r randomizer) Intn(n int) int {
	return r.Int() % n
}
func (r randomizer) Int63n(n int64) int64 {
	return r.Int63() % n
}
func (r randomizer) Uint63() uint64 {
	return new(maphash.Hash).Sum64()
}
func (r randomizer) Int63() int64 {
	u := int64(r.Uint63())
	if u < 0 {
		u *= -1
	}
	return u
}
func (r randomizer) Int() int {
	return int(r.Int63())
}
