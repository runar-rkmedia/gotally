package tallylogiccompaction

import (
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"strings"
	"testing"

	"github.com/gookit/color"
)

func TestWriteTripletAt(t *testing.T) {
	rnd := rand.NewSource(123)
	bytes := randomBytes(rnd, 30)
	{
		b := byte(0b101)
		tCount := maxTripletCount(len(bytes))
		for i := 0; i < tCount; i++ {
			// t.Logf("bytes before %08b", bytes)
			bytesBeforeStr := printb(bytes)
			atBefore := tripletAt(bytes, i)
			writeTripletAt(&bytes, i, b)
			atAfter := tripletAt(bytes, i)
			t.Logf("%d %03b\n", i, atAfter)
			if atAfter != b {
				bytesAfterStr := printb(bytes)
				t.Logf("byteas after %08b", bytes)
				t.Log("diff", highlightDiff(bytesBeforeStr, bytesAfterStr))
				t.Fatalf("At index %d should have written %03b, but it was read back as %03b (before: %03b)", i, b, atAfter, atBefore)
			}
		}
		// The error here can be ignored in this test, since we are producing random sets of bytes, which will probably not be correct
		tr, _ := byteSliceToTriplets(bytes)
		for i := 0; i < len(tr); i++ {
			if tr[i] != b {
				t.Log(printb(bytes))
				t.Errorf("Expected %03b at index %d, got  '%03b'\n %03b", b, i, tr[i], tr)
				break
			}
		}
	}
}
func TestAppendTripletAt(t *testing.T) {
	// return
	rnd := rand.NewSource(123)
	triplet := byte(0b101)
	for i := 0; i < 10; i++ {
		length := int(rnd.Int63())%30 + 1
		t.Run(fmt.Sprintf("Append length %d", length), func(t *testing.T) {

			bytes := randomBytes(rnd, length)
			t.Logf("Raw bytes before: %08b", bytes)

			tBefore, err := byteSliceToTriplets(bytes)
			if err != nil {
				return
			}

			tBefore = removeEmptyTripletsAtEnd(tBefore)
			lenBefore := len(tBefore)
			wroteAtIndex := appendTriplet(&bytes, triplet)
			t.Logf("RRaw bytes after: %08b (%d)", bytes, wroteAtIndex)
			tAfter, _ := byteSliceToTriplets(bytes)
			t.Logf("RRaw triplets after: %03b", tAfter)
			tAfter = removeEmptyTripletsAtEnd(tAfter)
			t.Logf("RRaw triplets after: %03b", tAfter)
			lenAfter := len(tAfter)
			if lenBefore != lenAfter-1 {
				t.Logf("before %03b", tBefore)
				t.Logf("after  %03b", tAfter)
				// t.Logf("after %08b", bytes)
				t.Errorf("Exptected there to be one more triplet after append, but the count was %d(%d) ; append(bytes(%d), %d)", lenAfter, lenBefore, len(bytes), triplet)
			}
			at := tripletAt(bytes, wroteAtIndex)
			if at != triplet {
				t.Errorf("Expected the last triplet after append(bytes, %d) to be %d, but it was %d", triplet, triplet, at)
			}
		})
	}

}
func TestByteSliceToTriplets(t *testing.T) {
	tests := []struct {
		name    string
		bytes   []byte
		want    []byte
		wantErr error
	}{
		// TODO: see if we can remove the last (up to two) 000-triplets that comes from padding
		{
			"Single byte",
			[]byte{0b010_110_00},
			[]byte{0b010, 0b110},
			nil,
		},
		{
			"Two bytes",
			[]byte{0b010_110_10, 0b1_111_000_0},
			[]byte{0b010, 0b110, 0b101, 0b111, 0b000},
			nil,
		},
		{
			"Two bytes variation",
			[]byte{0b011_101_01, 0b0_000_101_0},
			[]byte{0b011, 0b101, 0b010, 0b000, 0b101},
			nil,
		},
		{
			"Three bytes",
			[]byte{0b011_101_01, 0b0_000_101_1, 0b10_111_000},
			[]byte{0b011, 0b101, 0b010, 0b000, 0b101, 0b110, 0b111, 0b000},
			nil,
		},
		{
			"Four bytes",
			[]byte{0b011_101_01, 0b0_000_101_1, 0b10_111_100, 0b100_111_00},
			[]byte{0b011, 0b101, 0b010, 0b000, 0b101, 0b110, 0b111, 0b100, 0b100, 0b111},
			nil,
		},
		{
			"wantErr: invalid end of stream (ends with two bits to be read, but there is no next first bit)",
			[]byte{0b01110100, 0b10001001, 0b01010100, 0b00011010},
			[]byte{},
			ErrInvalidEnd,
		},
		{
			"wantErr: invalid end of stream (ends with two bits to be read, but there is no next first bit)",
			[]byte{0b11010101, 0b00011100, 0b10110011, 0b10101101, 0b11110010, 0b10101011, 0b01110000, 0b01100111},
			[]byte{},
			ErrInvalidEnd,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt.Println("********************************************************************************")
			fmt.Printf("** Start Read %s %s %08b\n", tt.name, printb(tt.bytes), tt.bytes)
			fmt.Println("********************************************************************************")
			got, err := byteSliceToTriplets(tt.bytes)
			if err != tt.wantErr {
				t.Errorf("expected error did not match: got %v wanted %v", err, tt.wantErr)
			}
			if err == nil {
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("readTriplets(%v) = \n%v, want \n%v", tt.bytes, got, tt.want)
					t.Errorf("readTriplets(%08b) = \n%03b, want \n%03b", tt.bytes, got, tt.want)
					bdiff(t, got, tt.want)
				}
			}
			for i := 0; i < len(got); i++ {
				triplet := got[i]
				atIndex := tripletAt(tt.bytes, i)
				fmt.Printf("at: %d %03b %03b\n", i, triplet, atIndex)
				if triplet != atIndex {
					t.Errorf("atIndex mismatch at index %d, expected %03b but got %03b", i, triplet, atIndex)
				}
			}
		})
	}
}
func TestTripletsToByteSlice(t *testing.T) {
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
			if got := tripletsToByteSlice(tt.bytes); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("fv writeTriplets(%v) = %v, want %v", tt.bytes, got, tt.want)
				t.Errorf("fp writeTriplets(%s) = \n%s want \n%s", printb(tt.bytes), printb(got), printb(tt.want))
				t.Errorf("fb writeTriplets(%08b) = %08b, want %08b", tt.bytes, got, tt.want)
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
func TestWriteReadBackTriplets(t *testing.T) {
	rnd := rand.NewSource(123)

	for n := 0; n < 100; n++ {
		length := int(rnd.Int63()%120) + 1
		t.Run(fmt.Sprintf("WriteReadBack %d length %d", n, length), func(t *testing.T) {
			bytes := randomBytes(rnd, length)
			t.Log("bytes", n, length, printb(bytes))
			fmt.Println("********************************************************************************")
			fmt.Printf("** Start Read %d %d %s %08b\n", n, length, printb(bytes), bytes)
			fmt.Println("********************************************************************************")
			triplets, err := byteSliceToTriplets(bytes)
			stringTriplets := strings.Split(printb(bytes), " ")
			lastStringTriplet := stringTriplets[len(stringTriplets)-1]
			if err != nil {
				if len(lastStringTriplet) != 3 && err == ErrInvalidEnd {
					// In this case, the bytes generated is invalid, and we are expecting an error
					return
				}
				t.Error("got unexpected err", err)
				return
			}
			readBytes := tripletsToByteSlice(triplets)
			t.Logf("written \nin bytes: %s\ntriplets: %03b\nback:     %s\n", printb(bytes), triplets, printb(readBytes))
			t.Logf("written \nin bytes: %v\nback:     %v\n", bytes, readBytes)
			t.Logf("written \nin bytes: %08b\nback:     %08b\n", bytes, readBytes)
			if !reflect.DeepEqual(readBytes, bytes) {
				t.Log("orig", printb(bytes))
				t.Log("read", printb(readBytes))
				t.Errorf("writing triplets and back to bytes mismatch \ngot:  %v, \nwant: %v", readBytes, bytes)
				bdiff(t, readBytes, bytes)
			}
			for i := 0; i < len(triplets); i++ {
				triplet := triplets[i]
				atIndex := tripletAt(bytes, i)
				fmt.Printf("at: %d %03b %03b\n", i, triplet, atIndex)
				if triplet != atIndex {
					t.Errorf("atIndex mismatch at index %d, expected %03b but got %03b", i, triplet, atIndex)
				}

			}
			tCount := maxTripletCount(len(bytes))
			if tCount != len(triplets) {
				t.Errorf("tripletCount(%d) mismatch, expected %d, but got %d", len(bytes), len(triplets), tCount)
			}
			trimmedTriplets := removeEmptyTripletsAtEnd(triplets)
			tCountExcludingPadding := tripletCount(bytes)
			if tCountExcludingPadding != len(trimmedTriplets) {
				t.Errorf("tripletCountExcludingPadding([%d bytes]) mismatch, expected %d, but got %d", len(trimmedTriplets), len(trimmedTriplets), tCountExcludingPadding)
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

// prints a byteslice as triplets. Mostly used for comparisons
func printb(b []byte) string {
	if len(b) == 0 {
		return "!!!Not bytes!! (len0)"
	}
	s := fmt.Sprintf("%08b", b)
	s = s[1 : len(s)-1]
	s = strings.ReplaceAll(s, " ", "")
	s2 := ""
	for i := 0; i < len(s); i++ {
		if i%3 == 0 {
			s2 += " "
		}
		s2 += string(s[i])
	}
	return s2
}
func randomBytes(rnd rand.Source, length int) []byte {
	bytes := make([]byte, rnd.Int63()%int64(length)+1)
	for i := 0; i < len(bytes); i++ {
		bytes[i] = byte(rnd.Int63())
	}
	return bytes
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
