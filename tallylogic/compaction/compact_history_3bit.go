package tallylogiccompaction

import (
	"fmt"

	"github.com/runar-rkmedia/gotally/tallylogic"
)

// Instructionms are normally 4-bits.
// instructiontable:
// 00-03:  SelectDirection: Selects the next cell in that direction from the previously selected one.
// 04:     ModePath: Enters Path-Mode, which is active until a non-SelectDirection-instruction follows, or we reach the end of the stream.
// 05-08:  SwipeDirection: Swipes in that direction
// 09:     Hint
// 10:     Undo
// 11:     Reserved for Swap
// 12:     Reserved for currently unknown feature
// 13:     Reserved for currently unknown feature
// 14:     Reserved for currently unknown feature
// 15:     Reserved for activating extra mode, in case we need more bits. When used without any following instructions, it is simply used as padding.

// TODO: make bitGroupModePath the zeroeth and the padding, since it cannot be used at the end anyway
const (
	bitGroupDirectionUp = iota
	bitGroupDirectionRight
	bitGroupDirectionDown
	bitGroupDirectionLeft
	bitGroupModePath
	bitGroupSwipeUp
	bitGroupSwipeRight
	bitGroupSwipeDown
	bitGroupSwipeLeft
	bitGroupHint
	bitGroupUndo
	bitGroupReservedSwap
	bitGroupReservedUnknown
	bitGroupReservedUnknown2
	bitGroupReservedUnknown3
	bitGroupPaddingOrModeExtraBitSizeIncrease
)

func NewCompactHistory(history tallylogic.Instruction) ([]byte, error) {

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

// TODO: use marshalbinary-interface for this, on the Instruction
func UnmarshalCompactHistory(bytes []byte, columns, rows int) (tallylogic.Instruction, error) {
	var history tallylogic.Instruction
	modePath := false
	var path []int
	nibbles := make([]byte, len(bytes)*2)
	for i, b := range bytes {
		// Splits the byte into two, so we get two nibbles
		nibbles[2*i] = b >> 4
		nibbles[2*i+1] = b << 4 >> 4
	}
	for i, nibble := range nibbles {
		if modePath {
			if len(path) == 0 {
				prev := nibbles[i-1]
				if prev == bitGroupModePath {
					continue
				}
				// combine the two next bitgroups into an int, which should be
				// the start of the path.
				if columns*rows > 256 {
					// TODO: increase this limit by checking how many bytes would be needed, and then reading in another bitGroup
					return history, fmt.Errorf("Not implemented! UnmarshalCompactHistory currently only supports boards with up to 256 cells.")
				}
				path = append(path, int(0|(prev<<4)|(nibble)))
				continue
			}
			prevPath := path[len(path)-1]
			noMatch := false
			switch nibble {
			case bitGroupDirectionUp:
				path = append(path, prevPath-columns)
			case bitGroupDirectionRight:
				path = append(path, prevPath+1)
			case bitGroupDirectionDown:
				path = append(path, prevPath+columns)
			case bitGroupDirectionLeft:
				path = append(path, prevPath-1)
			case bitGroupPaddingOrModeExtraBitSizeIncrease:
			default:
				noMatch = true
			}
			isLast := i == len(nibbles)-1
			if noMatch || isLast {
				history = append(history, path)
				path = []int{}
				modePath = false
			}
			if !noMatch {
				continue
			}
		}
		switch nibble {
		case bitGroupModePath:
			modePath = true
		case bitGroupSwipeUp:
			history = append(history, tallylogic.SwipeDirectionUp)
		case bitGroupSwipeRight:
			history = append(history, tallylogic.SwipeDirectionRight)
		case bitGroupSwipeDown:
			history = append(history, tallylogic.SwipeDirectionDown)
		case bitGroupSwipeLeft:
			history = append(history, tallylogic.SwipeDirectionLeft)
		case bitGroupPaddingOrModeExtraBitSizeIncrease:
			continue
		default:
			var modeString string
			switch {
			case modePath:
				modeString = "path"
			default:
				modeString = "regular"
			}
			return history, fmt.Errorf("Unhandled bitgroup %d at position %d in compact history for mode %q", nibble, i, modeString)
		}
	}
	// TOOD: read 4 bits at a time.
	return history, nil
}
