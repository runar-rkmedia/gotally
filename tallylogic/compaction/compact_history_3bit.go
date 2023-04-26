package tallylogiccompaction

import (
	"fmt"
	"strings"

	"github.com/runar-rkmedia/gotally/tallylogic"
)

const (
	bModePath = iota
	bModeHelpers
	bSwipeUp
	bSwipeRight
	bSwipeDown
	bSwipeLeft
)

const (
	bitgroupModePathToDefault = iota
	bitgroupModePathToHelpers
	bitgroupModePathUp
	bitgroupModePathRight
	bitgroupModePathDown
	bitgroupModePathLeft
)
const (
	bitgroupModeHelperHint = iota
	bitgroupModeHelperUndo
	bitgroupModeHelperSwap
)

func NewCompact3History(history tallylogic.Instruction) ([]byte, error) {

	b := []byte{}
	// index := 0
	for _, ins := range history {
		kind := tallylogic.GetInstructionType(ins)
		switch kind {
		case tallylogic.InstructionTypeSwipe:

		case tallylogic.InstructionTypeCombinePath:
		default:
			return nil, fmt.Errorf("Unhandled instuctionkind %v", kind)
		}

	}

	return b, fmt.Errorf("Not implemented")
}

// returns triplets from a byteslice
func writeTriplets(triplets []byte) []byte {
	length := len(triplets)*3/8 + 1
	bytes := make([]byte, length)

	for i := 0; i < len(triplets); i++ {
		ii := i % 8
		offset := 8 - (ii % 3 * 3) - 3 - ((ii / 3) % 3)
		j := (i + i/8) / 3
		if i > 0 && i%3 == 0 {
			jj := j % 3

			switch jj {
			case 1:
				// This byte should have the left-most bit from the previous triplet
				bytes[j] |= triplets[i-1] << 7
			case 2:
				// This byte should have the two left-most bit from the previous triplet
				bytes[j] |= triplets[i-1] << 6
			case 0:
			default:
				panic(fmt.Sprintf("unexpected modulo for triplet %d at i %d", jj, i))
			}
		}
		if offset > 0 {
			bytes[j] |= triplets[i] << offset
		} else {
			bytes[j] |= triplets[i] >> (offset * -1)
		}
		fmt.Println(printb(bytes))
	}
	fmt.Printf("write \n%03b\n", triplets)
	fmt.Printf("%s", printb(bytes))
	return bytes
}
func printb(b []byte) string {
	if len(b) == 0 {
		return "Not bytes"
	}
	s := fmt.Sprintf("%08b", b)
	s = s[1 : len(s)-1]
	// s = strings.TrimPrefix(s, "0")
	s = strings.ReplaceAll(s, " ", "")
	s2 := "\n"
	for i := 0; i < len(s); i++ {
		if i%3 == 0 {
			s2 += " "
		}
		s2 += string(s[i])
	}
	s2 += "\n"
	// for i := 0; i < len(s); i += 3 {
	// 	if i%3 == 0 {
	// 		s2 += " "
	// 	}
	// 	s2 += fmt.Sprintf("% 3d", i/3)
	// }
	return s2
}

// converts triplets into a byte-slice
func readTriplets(bytes []byte) []byte {
	triplets := []byte{}
	// i is the triplets-count currently written
	i := 0
	for j, b := range bytes {
		ii := i % 8
		fmt.Println("byte at index", j, "i", i, "ii", ii)

		// helper function while developing, remove
		p := func() {
			ii := i % 8
			offset := 8 - (ii % 3 * 3) - 3 - ((ii / 3) % 3)
			fmt.Printf("triplet i: %d j: %d o:%02d b: %08b %03b |", i, j, offset, b, triplets[len(triplets)-1])
			for _, v := range triplets {
				fmt.Printf(" %03b", v)
			}
			fmt.Printf("\n")

		}
		// Read from this index in the current byte (and perhaps from next too)
		switch ii {
		case 0:
			// Read two triplets from left to right.
			// Remainder: 2 bits
			triplets = append(triplets, b>>5)
			fmt.Printf("first %03b\n", triplets)
			i++
			p()
			triplets = append(triplets, b<<3>>5)
			i++
			p()
		case 2:
			// Read two bits from previous at right and one bits from the current byte at left
			// Then, Read two triplets.
			// Remainder: 1 bit.
			prevLastTwo := bytes[j-1] << 6 >> 5
			first := b >> 7
			triplets = append(triplets, prevLastTwo|first)
			i++
			p()
			fmt.Printf("chinkolo %08b %08b\n", bytes[j-1], b)
			fmt.Printf("prevlast %03b %03b %03b\n", prevLastTwo, first, triplets[len(triplets)-1])
			triplets = append(triplets, b<<1>>5)
			i++
			p()
			triplets = append(triplets, b<<4>>5)
			p()
			i++
		case 5:
			// Read one bit from previous at right and two bits from the current byte at left
			// Then, Read two triplets.
			// Remainder: 0 bit.
			prevLast := bytes[j-1] << 7 >> 5
			firstTwoBit := b >> 6
			triplets = append(triplets, prevLast|firstTwoBit)
			i++
			fmt.Printf("4 chinkolo %08b %08b\n", bytes[j-1], b)
			fmt.Printf("4 prevlast %03b %03b %03b\n", prevLast, firstTwoBit, triplets[len(triplets)-1])
			p()
			triplets = append(triplets, b<<2>>5)
			p()
			i++
			triplets = append(triplets, b<<5>>5)
			p()
			i++

		default:
			fmt.Println("whoops input", ii, i, j)
			panic("whoops")
		}

	}
	fmt.Printf("read %s\n", printb(bytes))
	fmt.Printf("%03b\n", triplets)

	return triplets
}

// reads triplets from a binary-range
func _readTriplets(bytes []byte) []byte {
	// length := len(bytes) * 8 / 3
	// triplets := make([]byte, length)
	triplets := []byte{}
	// fmt.Printf("length %d -> %d\n", len(bytes), length)
	i := 0
	tripletOffset := 0
	var tripletMask byte
	for _, b := range bytes {
		switch tripletOffset {
		case 0:
			triplets = append(triplets, b>>5)
			i++
			triplets = append(triplets, b<<3>>5)
			i++
			tripletMask = b << 6 >> 5
			// fmt.Printf("b %d %08b (%d) -> %08b-%08b | %08b\n", j, b, tripletOffset, triplets[i-1], triplets[i], tripletMask)
			// fmt.Printf("b %d %08b (%d) -> %03b-%03b | %03b\n", j, b, tripletOffset, triplets[i-2], triplets[i-1], tripletMask)
			tripletOffset = 1
		case 1:
			firstBit := b >> 7
			i++
			triplets = append(triplets, firstBit|tripletMask)
			triplets = append(triplets, b<<1>>5)
			i++
			triplets = append(triplets, b<<4>>5)
			i++
			// fmt.Printf("First: %03b %03b %03b\n", firstBit, tripletMask, firstBit|tripletMask)
			// fmt.Printf("b %d %08b (%d) -> %03b %03b | %03b\n", j, b, tripletOffset, triplets[i-2], triplets[i-1], tripletMask)
			tripletMask = b << 7 >> 5
			// fmt.Printf("\tmask(1) from %08b %03b\n", b, tripletMask)
			tripletOffset = 2
		case 2:
			firstTwoBits := b >> 5
			i++
			triplets = append(triplets, firstTwoBits|tripletMask)
			triplets = append(triplets, b<<3>>5)
			i++
			// triplets = append(triplets, b<<4>>5)
			// i++
			// fmt.Printf("FirstTwo from %08b: %03b %03b %03b\n", b, firstTwoBits, tripletMask, firstTwoBits|tripletMask)
			// fmt.Printf("b %d %08b (%d) -> %03b %03b | %03b\n", j, b, tripletOffset, triplets[i-2], triplets[i-1], tripletMask)
			tripletOffset = 1
		}
	}
	// fmt.Printf("real length(%d):  %d -> %d\n", length, len(bytes)*8, len(triplets))
	return triplets
}

// // TODO: use marshalbinary-interface for this, on the Instruction
// func UnmarshalCompactHistory3(bytes []byte, columns, rows int) (tallylogic.Instruction, error) {
// 	var history tallylogic.Instruction
// 	modePath := false
// 	var path []int
// 	triplets := make([]byte, len(bytes)*2)
// 	tripletOffset := 0
// 	// var tripletMask byte
// 	for i, b := range bytes {
// 		switch tripletOffset {
// 		case 0:
// 			triplets[2*i] = b >> 6
// 			triplets[2*i+1] = b << 6 >> 3
//
// 		default:
// 			return history, fmt.Errorf("Unhandled tripletOffset: %d", tripletOffset)
// 		}
// 		triplets[2*i] = b >> 4
// 		triplets[2*i+1] = b << 4 >> 4
// 	}
// 	for i, nibble := range triplets {
// 		if modePath {
// 			if len(path) == 0 {
// 				prev := triplets[i-1]
// 				if prev == ModePath {
// 					continue
// 				}
// 				// combine the two next bs into an int, which should be
// 				// the start of the path.
// 				if columns*rows > 256 {
// 					// TODO: increase this limit by checking how many bytes would be needed, and then reading in another b
// 					return history, fmt.Errorf("Not implemented! UnmarshalCompactHistory currently only supports boards with up to 256 cells.")
// 				}
// 				path = append(path, int(0|(prev<<4)|(nibble)))
// 				continue
// 			}
// 			prevPath := path[len(path)-1]
// 			noMatch := false
// 			switch nibble {
// 			case bSwipeUp:
// 				path = append(path, prevPath-columns)
// 			case bDirectionRight:
// 				path = append(path, prevPath+1)
// 			case bDirectionDown:
// 				path = append(path, prevPath+columns)
// 			case bDirectionLeft:
// 				path = append(path, prevPath-1)
// 			case bPaddingOrModeExtraBitSizeIncrease:
// 			default:
// 				noMatch = true
// 			}
// 			isLast := i == len(triplets)-1
// 			if noMatch || isLast {
// 				history = append(history, path)
// 				path = []int{}
// 				modePath = false
// 			}
// 			if !noMatch {
// 				continue
// 			}
// 		}
// 		switch nibble {
// 		case bModePath:
// 			modePath = true
// 		case bSwipeUp:
// 			history = append(history, tallylogic.SwipeDirectionUp)
// 		case bSwipeRight:
// 			history = append(history, tallylogic.SwipeDirectionRight)
// 		case bSwipeDown:
// 			history = append(history, tallylogic.SwipeDirectionDown)
// 		case bSwipeLeft:
// 			history = append(history, tallylogic.SwipeDirectionLeft)
// 		case bPaddingOrModeExtraBitSizeIncrease:
// 			continue
// 		default:
// 			var modeString string
// 			switch {
// 			case modePath:
// 				modeString = "path"
// 			default:
// 				modeString = "regular"
// 			}
// 			return history, fmt.Errorf("Unhandled b %d at position %d in compact history for mode %q", nibble, i, modeString)
// 		}
// 	}
// 	// TOOD: read 4 bits at a time.
// 	return history, nil
// }
