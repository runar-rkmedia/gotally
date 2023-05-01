package triplets

import (
	"errors"
	"fmt"
	"math"
)

// The table below displays a visual representation
// of a bytes, and the order for reading each triplet
// this is only to help understand what is happening.
// each [] represents a triplet
//
// bit-offset:  7  6  5  4  3  2  1  0
// bit-offset:  0  1  2  3  4  5  6  7
// -----------------------------------
// byte 0:     [0  0  0][0  0  1][0  1
// byte 1:      0][0  1  1][1  0  0][ 1
// byte 2:      0  1][1  1  0][1  1  1]
//
// from this we can see:
// triplet 0: starts at byte 0, offset 0-2 (7-5)
// triplet 1: starts at byte 0, offset 3-5 (4-2)
// triplet 2: starts at byte 0, offset 6 (1)
//              ends at byte 1, offset 0 (7)
// triplet 3: starts at byte 1, offset 1-3 (6-4)
// triplet 4: starts at byte 1, offset 4-6 (3-1)
// triplet 5: starts at byte 1, offset 7 (0)
//              ends at byte 2, offset 1 (6)
// triplet 6: starts at byte 2, offset 2-4 (5-3)
// triplet 7: starts at byte 2, offset 5-7 (2-0)

func tripletsToByteSlice(triplets []byte) []byte {

	tripletCount := len(triplets)
	if tripletCount == 0 {
		return nil
	}
	lengthf := float64(len(triplets)) * 3 / 8
	byteCount := int(math.Ceil(lengthf))
	bytes := make([]byte, byteCount)

	byteIndex := 0

	for it := 0; it < tripletCount; it++ {
		switch it % 8 {
		case 0:
			bytes[byteIndex] |= triplets[it] << 5
		case 1:
			bytes[byteIndex] |= triplets[it] << 2
		case 2:
			bytes[byteIndex] |= triplets[it] >> 1
			byteIndex++
		case 3:
			bytes[byteIndex] |= triplets[it-1] << 7
			bytes[byteIndex] |= triplets[it] << 4
		case 4:
			bytes[byteIndex] |= triplets[it] << 1
		case 5:
			bytes[byteIndex] |= triplets[it] >> 2
			byteIndex++
		case 6:
			bytes[byteIndex] |= triplets[it-1] << 6
			bytes[byteIndex] |= triplets[it] << 3
		case 7:
			bytes[byteIndex] |= triplets[it]
			byteIndex++
		}
	}

	return bytes
}

var (
	ErrInvalidEnd = errors.New("Invalid end of []byte (expected 0)")
)

func byteSliceToTriplets(bytes []byte) ([]byte, error) {
	// i is the triplets-count currently written
	i := 0
	byteCount := len(bytes)
	triplets := make([]byte, maxTripletCount(byteCount))
	for j, b := range bytes {
		ii := i % 8

		// Read from this index in the current byte (and perhaps from next too)
		switch ii {
		case 0:
			// Read two triplets from left to right.
			// Remainder: 2 bits
			triplets[i] = b >> 5
			i++
			triplets[i] = b << 3 >> 5
			i++
			if j == byteCount-1 {
				lastTwo := bytes[j] << 6 >> 5
				if lastTwo != 0b0 {
					return triplets, ErrInvalidEnd
				}
			}
		case 2:
			// Read two bits from previous at right and one bits from the current byte at left
			// Then, Read two triplets.
			// Remainder: 1 bit.
			triplets[i] = bytes[j-1]<<6>>5 | b>>7
			i++
			triplets[i] = b << 1 >> 5
			i++
			triplets[i] = b << 4 >> 5
			i++
			if j == byteCount-1 {
				last := bytes[j] << 7 >> 5
				if last != 0b0 {
					return triplets, ErrInvalidEnd
				}
			}
		case 5:
			// Read one bit from previous at right and two bits from the current byte at left
			// Then, Read two triplets.
			// Remainder: 0 bit.
			triplets[i] = bytes[j-1]<<7>>5 | b>>6
			i++
			triplets[i] = b << 2 >> 5
			i++
			triplets[i] = b << 5 >> 5
			i++

		default:
			panic(fmt.Sprintf("unexpected offset during readTriplets ii:%d, i: %d, j: %d", ii, i, j))
		}

	}
	return triplets, nil
}

// returns a triplet at the given position
// the bounds-check is expected to be performed in advance
func tripletAt(bytes []byte, index int) byte {
	mod := index % 8
	bMultiple := index / 8 * 3
	switch mod {
	case 0:
		return bytes[bMultiple] >> 5
	case 1:
		return bytes[bMultiple] << 3 >> 5
	case 2:
		return bytes[bMultiple]<<6>>5 | bytes[bMultiple+1]>>7
	case 3:
		return bytes[bMultiple+1] << 1 >> 5
	case 4:
		return bytes[bMultiple+1] << 4 >> 5
	case 5:
		return bytes[bMultiple+1]<<7>>5 | bytes[bMultiple+2]>>6
	case 6:
		return bytes[bMultiple+2] << 2 >> 5
	case 7:
		return bytes[bMultiple+2] << 5 >> 5
	}
	return 0
}

// appends a triplet, ignoring any existing padding
func appendTriplet(bytes *[]byte, triplets ...byte) int {
	// bl := len(*bytes)
	l := tripletCount(*bytes)
	byteCount := int(math.Ceil(float64(l) * 3 / 8))
	newByteCount := int(math.Ceil(float64(l+len(triplets)) * 3 / 8))

	bytesNeeded := newByteCount - byteCount
	for i := 0; i < bytesNeeded; i++ {
		*bytes = append(*bytes, 0)
	}

	for i := 0; i < len(triplets); i++ {
		writeTripletAt(bytes, l+i, triplets[i])
	}
	return l
}

// Writes a triplet at the triplet-index into the compact []byte
func writeTripletAt(bytes *[]byte, index int, triplet byte) {
	mod := index % 8
	bMultiple := index / 8 * 3
	if index == 17 {
	}
	switch mod {
	case 0:
		(*bytes)[bMultiple] ^= ((*bytes)[bMultiple] ^ triplet<<5) & 0b1110_0000
	case 1:
		(*bytes)[bMultiple] ^= ((*bytes)[bMultiple] ^ triplet<<2) & 0b0001_1100
	case 2:
		(*bytes)[bMultiple] ^= ((*bytes)[bMultiple] ^ triplet>>1) & 0b0000_0011
		(*bytes)[bMultiple+1] ^= ((*bytes)[bMultiple+1] ^ triplet<<7) & 0b1000_0000
	case 3:
		(*bytes)[bMultiple+1] ^= ((*bytes)[bMultiple+1] ^ triplet<<4) & 0b0111_0000
	case 4:
		(*bytes)[bMultiple+1] ^= ((*bytes)[bMultiple+1] ^ triplet<<1) & 0b0000_1110
	case 5:
		(*bytes)[bMultiple+1] ^= ((*bytes)[bMultiple+1] ^ triplet>>2) & 0b0000_0001
		(*bytes)[bMultiple+2] ^= ((*bytes)[bMultiple+2] ^ triplet<<6) & 0b1100_0000
	case 6:
		(*bytes)[bMultiple+2] ^= ((*bytes)[bMultiple+2] ^ triplet<<3) & 0b0011_1000
	case 7:
		(*bytes)[bMultiple+2] ^= ((*bytes)[bMultiple+2] ^ triplet) & 0b0000_0111
	}
}

// returns the maximum triplet-count that can be stored in a []byte of this length
func maxTripletCount(byteCount int) int {
	return byteCount * 8 / 3
}
func removeEmptyTripletsAtEnd(triplets []byte) []byte {
	if len(triplets) == 0 {
		return triplets
	}
	r := 0
	for i := len(triplets) - 1; i >= 0; i-- {
		if triplets[i] == 0 {
			if r > 2 {
				panic(fmt.Sprintf("too many empty triplets: %d", r))
			}
			triplets = triplets[:i]
			r++
		} else {
			return triplets
		}
	}
	return triplets
}

// returns the triplet-count for a []byte, ignoring any 0-triplets at the end.
func tripletCount(bytes []byte) int {
	l := len(bytes) * 8 / 3
	for l > 0 && tripletAt(bytes, l-1) == 0 {
		l--
	}
	return l
}

type CompactTriplets []byte

func NewCompactTriplets(b []byte) CompactTriplets {
	return CompactTriplets(b)
}

func (c *CompactTriplets) Append(triplets ...byte) int {
	b := []byte(*c)
	i := appendTriplet(&b, triplets...)
	*c = CompactTriplets(b)

	return i
}
func (c *CompactTriplets) WriteAt(index int, triplet byte) {
	b := []byte(*c)
	writeTripletAt(&b, index, triplet)
}
func (c CompactTriplets) At(index int) byte {
	b := []byte(c)
	return tripletAt(b, index)
}
func (c CompactTriplets) Length() int {
	b := []byte(c)
	return tripletCount(b)
}
func (c CompactTriplets) Size() int {
	return len(c)
}
func (c CompactTriplets) Unpack() ([]byte, error) {
	return byteSliceToTriplets(c)
}
