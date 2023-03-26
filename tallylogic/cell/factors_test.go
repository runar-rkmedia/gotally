package cell

import (
	"fmt"
	"testing"
	"time"

	"github.com/MarvinJWendt/testza"
	"github.com/go-test/deep"
)

func init() {
	testza.SetShowStartupMessage(false)
}

func Test_findPrimes(t *testing.T) {
	tests := []struct {
		n    uint64
		want []uint64
	}{
		{0, []uint64{}},
		{1, []uint64{1}},
		{2, []uint64{2}},
		{3, []uint64{3}},
		{4, []uint64{2, 2}},
		{5, []uint64{5}},
		{6, []uint64{2, 3}},
		{7, []uint64{7}},
		{8, []uint64{2, 2, 2}},
		{9, []uint64{3, 3}},
		{10, []uint64{2, 5}},
		{11, []uint64{11}},
		{12, []uint64{2, 2, 3}},
		{13, []uint64{13}},
		{14, []uint64{2, 7}},
		{15, []uint64{3, 5}},
		{16, []uint64{2, 2, 2, 2}},
		{17, []uint64{17}},
		{18, []uint64{2, 3, 3}},
		{19, []uint64{19}},
		{20, []uint64{2, 2, 5}},
		{96, []uint64{2, 2, 2, 2, 2, 3}},
		{256, []uint64{2, 2, 2, 2, 2, 2, 2, 2}},
		{2038, []uint64{2, 1019}},
		{2039, []uint64{2039}},
		{2041, []uint64{13, 157}},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("Should return correctly for input %d", tt.n), func(t *testing.T) {

			// in case of an infinite loop
			var primes []uint64
			var primes2 []uint64
			testza.AssertCompletesIn(t, time.Millisecond*50, func() {
				primes = primeFactors(tt.n)
				primes2 = _primeFactors(tt.n)
			})

			if diff := deep.Equal(primes, tt.want); diff != nil {
				t.Logf("Number of created primes: %d (%d)", primeCount, len(listOfPrimes))
				t.Errorf("primefactors(%d) = %v, want %v\ndiff: %#v", tt.n, primes, tt.want, diff)
			}
			if diff := deep.Equal(primes, primes2); diff != nil {
				t.Logf("Number of created primes: %d (%d)", primeCount, len(listOfPrimes))
				t.Errorf("_primefactors(%d) = primeFactors(%d) =  %v, want %v\ndiff: %#v", tt.n, tt.n, primes, primes2, diff)
			}
		})
	}
}
func Test_createMorePrimes(t *testing.T) {
	tests := []struct {
		name  string
		limit uint
		want  []uint64
	}{
		{
			"Should create some primes",
			10,
			[]uint64{2, 3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37, 41},
		},
	}
	for _, tt := range tests {
		listOfPrimes = []uint64{2, 3}
		t.Run(tt.name, func(t *testing.T) {

			// in case of an infinite loop
			testza.AssertCompletesIn(t, time.Second, func() { createMorePrimes(tt.limit) })

			if diff := deep.Equal(listOfPrimes, tt.want); diff != nil {
				t.Errorf("getNeededCellsHighestForm() = %v, want %v\ndiff: %#v", listOfPrimes, tt.want, diff)
			}
		})
	}
}
