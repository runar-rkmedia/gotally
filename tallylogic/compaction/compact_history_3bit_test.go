package tallylogiccompaction

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"testing"

	"github.com/gookit/color"
)

func TestWriteReadBackTriplets(t *testing.T) {
	rnd := rand.NewSource(123)

	b := bytes.NewBuffer(nil)
	length := int(rnd.Int63() % 12)
	for i := 0; i < length; i++ {
		n := rnd.Int63()
		b.WriteByte(byte(n))
	}
	bytes := b.Bytes()
	t.Log("bytes", printb(bytes))
	triplets := readTriplets(bytes)
	readBytes := writeTriplets(triplets)
	t.Logf("written%s%03b%s", printb(bytes), triplets, printb(readBytes))
	if !reflect.DeepEqual(readBytes, bytes) {
		t.Log(printb(bytes))
		t.Log(printb(readBytes))
		t.Errorf("writing triplets and back to bytes mismatch \ngot:  %v, \nwant: %v", readBytes, bytes)
		bdiff(t, readBytes, bytes)
	}
}
func TestReadTriplets(t *testing.T) {
	tests := []struct {
		name  string
		bytes []byte
		want  []byte
	}{
		// TODO: see if we can remove the last (up to two) 000-triplets that comes from padding
		{
			"Single byte",
			[]byte{0b010_110_00},
			[]byte{0b010, 0b110},
		},
		{
			"Two bytes",
			[]byte{0b010_110_10, 0b1_111_000_0},
			[]byte{0b010, 0b110, 0b101, 0b111, 0b000},
		},
		{
			"Two bytes variation",
			[]byte{0b011_101_01, 0b0_000_101_0},
			[]byte{0b011, 0b101, 0b010, 0b000, 0b101},
		},
		{
			"Three bytes",
			[]byte{0b011_101_01, 0b0_000_101_1, 0b10_111_000},
			[]byte{0b011, 0b101, 0b010, 0b000, 0b101, 0b110, 0b111, 0b000},
		},
		{
			"Four bytes",
			[]byte{0b011_101_01, 0b0_000_101_1, 0b10_111_100, 0b100_111_10},
			[]byte{0b011, 0b101, 0b010, 0b000, 0b101, 0b110, 0b111, 0b100, 0b100, 0b111},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt.Println("********************************************************************************")
			fmt.Printf("** Start Read %s %s\n", tt.name, printb(tt.bytes))
			fmt.Println("********************************************************************************")
			if got := readTriplets(tt.bytes); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readTriplets(%v) = \n%v, want \n%v", tt.bytes, got, tt.want)
				t.Errorf("readTriplets(%08b) = \n%03b, want \n%03b", tt.bytes, got, tt.want)
				bdiff(t, got, tt.want)
			}
		})
	}
}
func bdiff(t *testing.T, got, want []byte) {
	if len(got) != len(want) {
		t.Errorf("Length mismatch got %d, want %d ", len(got), len(want))
	}
	for i := 0; i < int(math.Min(float64(len(want)), float64(len(got)))); i++ {
		w := want[i]
		g := got[i]
		if w != g {
			diff := highlightDiff(fmt.Sprintf("%08b", w), fmt.Sprintf("%08b", g))
			t.Errorf("index %d %08b != %08b %s", i, g, w, diff)
		}

	}

}
func TestWriteTriplets(t *testing.T) {
	tests := []struct {
		name  string
		bytes []byte
		want  []byte
	}{
		{
			"Single byte",
			[]byte{0b010, 0b110},
			[]byte{0b010_110_00},
		},
		{
			"Two bytes",
			[]byte{
				0b010, 0b110, 0b101,
				0b111, 0b000,
			},
			[]byte{
				0b010_110_10,
				0b1_111_000_0},
		},
		{
			"Two bytes 2",
			[]byte{
				0b011, 0b101, 0b010,
				0b000, 0b101,
			},
			[]byte{
				0b011_101_01,
				0b0_000_101_0,
			},
		},
		{
			"Three bytes",
			[]byte{
				0b011, 0b101, 0b010,
				0b000, 0b101, 0b110,
				0b111,
			},
			[]byte{
				0b011_101_01,
				0b0_000_101_1,
				0b10_111_000,
			},
		},
		{
			"Four bytes",
			[]byte{
				0b011, 0b101, 0b010,
				0b000, 0b101, 0b110,
				0b111, 0b101, 0b001,
				0b111,
			},
			[]byte{
				0b011_101_01,
				0b0_000_101_1,
				0b10_111_101,
				0b001_111_00,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := writeTriplets(tt.bytes); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("writeTriplets(%v) = %v, want %v", tt.bytes, got, tt.want)
				t.Errorf("writeTriplets(%08b) = %08b, want %08b", tt.bytes, got, tt.want)
				for i, w := range tt.want {
					if w != got[i] {
						diff := highlightDiff(fmt.Sprintf("%08b", w), fmt.Sprintf("%08b", got[i]))
						t.Errorf("index %d %08b != %08b %s", i, got[i], w, diff)
					}

				}
			}
		})
	}
}
func highlightDiff(before, after string) string {
	s := ""
	for i := range before {
		if before[i] != after[i] {
			s += color.Blue.Render(string(after[i]))
		} else {
			s += color.Yellow.Render(string(after[i]))

		}
	}
	return fmt.Sprintf("%s %s into %s", s, before, after)
}
