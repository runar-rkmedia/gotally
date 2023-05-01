package tallylogic

import (
	"fmt"
	"math"
	"math/bits"
	"strconv"
	"strings"

	"github.com/runar-rkmedia/gotally/triplets"
)

type history = byte

const (
	// Default Mode
	bModePath history = iota
	bModeHelpers
	bSwipeUp
	bSwipeRight
	bSwipeDown
	bSwipeLeft
)

const (
	// Mode Path. Must always end with PathEnd
	bitgroupModePathUp history = iota
	bitgroupModePathRight
	bitgroupModePathDown
	bitgroupModePathLeft
	// We have some extra bits available here, so we can do a combinator
	bitgroupModePathUpEnd
	bitgroupModePathRightEnd
	bitgroupModePathDownEnd
	bitgroupModePathLeftEnd
)
const (
	// Mode Helper. After any of these, the mode is returned to Default
	bitgroupModeHelperVoid history = iota
	bitgroupModeHelperHint
	bitgroupModeHelperUndo
	bitgroupModeHelperSwap
)
const (
	modePath    = "path"
	modeDefault = "default"
	modeHelper  = "helper"
)

type helper string

const (
	helperHint helper = "Hint"
	helperUndo helper = "Undo"
	helperSwap helper = "Swap"
)

type CompactHistory struct {
	c                        triplets.CompactTriplets
	gameRows, gameColumns    int
	bitsUsedForPathIndex     int
	tripletsUsedForPathIndex int
}

/*
	CompactHistory is a compact representation of the History of a Game.
	This history includes an ordered list of every Swipe, Cell-paths combines and Helpers use performed.
	The reason for using this history is to:
	- Support Undo
	- Rerun a game, to show how it was done.

	Infinite games are in practice not infinite, but can be many thousand of instructions long.
	Example:
	- 600 * Swipe: 1 byte
	- 50 * Helper: 1 byte
	- 350 * Path:  8 byte per path (avg 5) = 40byte
	Sum = 60 + 50 bytes + 350 bytes * 50 = 44000 bytes = 44kb
	In contrast to this compact representation:

	# Naive implementation

| ------- | ---------------- | ------------------------------------------------------------------- |
| Swipe   | 1 byte           |                                                                     |
| Helper  | 1 byte           |                                                                     |
| Path    | 5 byte           | `[]uint8`, 1 byte per path-index. (supports board-sizes up to 256), |
| **Sum** | **2400 bytes**   | **600 + 50 + 350 x 5**                                              |

# Compact implementation

|         | Size-requirement | Notes                                                                                           |
| ------- | ---------------- | ----------------------------------------------------------------------------------------------- |
| Swipe   | 3 bits           | 3 bits for value                                                                                |
| Helper  | 6 bits           | 3 bits for header start, 3 bits for value                                                       |
| Path    | 21 bits          | 3 bit for header start. <br>6 bits for first path-index, <br>3 bits for rest of the indexes<br> |
| **Sum** | **1181 bytes**  | **(600 x 3 + 50 x 3 + 350 x 21 )/ 8**                                                           |
*/
func NewCompactHistory(gameColumns, gameRows int) CompactHistory {
	boardSize := gameRows * gameColumns
	bitsUsedForPathIndex := bits.Len(uint(boardSize))
	tripletsUsedForPathIndex := int(math.Ceil(float64(bitsUsedForPathIndex) / 3))
	switch tripletsUsedForPathIndex {
	case 1, 2, 3:
	default:
		// There are two real places that should implement support for a higher triplet-size
		panic(fmt.Sprintf("NotImplemeted:NewCompactHistory tripletsUsedForPathIndex=%d for boardSize %d. BitsNeeded: %d", tripletsUsedForPathIndex, boardSize, bitsUsedForPathIndex))

	}
	return CompactHistory{
		[]byte{},
		gameRows,
		gameColumns,
		bitsUsedForPathIndex,
		tripletsUsedForPathIndex,
	}
}

func (c *CompactHistory) AddHint() {
	c.c.Append(bModeHelpers, bitgroupModeHelperHint)
}
func (c *CompactHistory) AddSwap() {
	c.c.Append(bModeHelpers, bitgroupModeHelperSwap)
}
func (c *CompactHistory) AddUndo() {
	c.c.Append(bModeHelpers, bitgroupModeHelperUndo)
}
func (c *CompactHistory) AddSwipe(dir SwipeDirection) {
	switch dir {
	case SwipeDirectionUp:
		c.c.Append(bSwipeUp)
	case SwipeDirectionRight:
		c.c.Append(bSwipeRight)
	case SwipeDirectionDown:
		c.c.Append(bSwipeDown)
	case SwipeDirectionLeft:
		c.c.Append(bSwipeLeft)
	}
}
func (c *CompactHistory) AddPath(path []int) error {
	length := len(path)
	if length < 2 {
		return fmt.Errorf("Path must be of at least of length 2")
	}
	// Start with a 0-byte(bModePath),
	// followed by the first index path as triplet-count defined by c.tripletsUsedForPathIndex
	toAppend := make([]byte, length+c.tripletsUsedForPathIndex)

	first := byte(path[0])
	switch c.tripletsUsedForPathIndex {
	case 1:
		toAppend[1] = first
	case 2:
		toAppend[1] = first & 0b00111000 >> 3
		toAppend[2] = first & 0b00000111
	case 3:
		toAppend[1] = first & 0b11000000 >> 6
		toAppend[2] = first & 0b00111000 >> 3
		toAppend[3] = first & 0b00000111
	default:
		panic(fmt.Sprintf("NotImplemeted:CompactHistory:AddPath tripletsUsedForPathIndex=%d (%d)", c.tripletsUsedForPathIndex, c.bitsUsedForPathIndex))
	}

	for i := 1; i < length; i++ {
		toAppend[i+c.tripletsUsedForPathIndex] = combinePathRelative(path[i-1], path[i])
		if i == length-1 {
			toAppend[i+c.tripletsUsedForPathIndex] += 4
		}
	}
	c.c.Append(toAppend...)
	return nil
}

func (c *CompactHistory) Describe() string {
	s := strings.Builder{}
	err := c.Iterate(
		func(dir SwipeDirection, i int) {
			switch dir {
			case SwipeDirectionUp:
				s.WriteString("U;")
			case SwipeDirectionRight:
				s.WriteString("R;")
			case SwipeDirectionDown:
				s.WriteString("D;")
			case SwipeDirectionLeft:
				s.WriteString("L;")
			}
		},
		func(path []int, i int) {
			l := len(path)
			for i := 0; i < l; i++ {
				s.WriteString(strconv.FormatInt(int64(path[i]), 10))
				if i < l-1 {
					s.WriteString(",")
				}
			}
			s.WriteString(";")
		},
		func(helper helper, i int) {
			switch helper {
			case helperHint:
				s.WriteString("H;")
			case helperSwap:
				s.WriteString("S;")
			case helperUndo:
				s.WriteString("Z;")
			}
		},
	)
	if err != nil {
		return fmt.Errorf("failed to describe: %w after %s", err, s.String()).Error()
	}
	return s.String()

}
func name(mode string, history history) string {
	switch mode {
	case modeDefault:
		switch history {
		case bModePath:
			return "Default: Mode -> Path"
		case bModeHelpers:
			return "Default: Mode -> Helpers"
		case bSwipeUp:
			return "Default: Swipe Up"
		case bSwipeRight:
			return "Default: Swipe Right"
		case bSwipeDown:
			return "Default: Swipe Down"
		case bSwipeLeft:
			return "Default: Swipe Left"
		}
	case modeHelper:
		switch history {
		case bitgroupModeHelperHint:
			return "Helper: Hint"
		case bitgroupModeHelperSwap:
			return "Helper: Swap"
		case bitgroupModeHelperUndo:
			return "Helper: Swap"
		}
	case modePath:
		return fmt.Sprintf("Path: %d", history)

	}
	return fmt.Sprintf("Unmapped history in mode '%s': %d", mode, history)
}
func (c *CompactHistory) Iterate(
	onSwipe func(dir SwipeDirection, i int),
	onCombinePath func(path []int, i int),
	onHelper func(helper helper, i int),
) error {
	l := c.c.Length()
	mode := modeDefault
	var j int
	path := []int{}
	for i := 0; i < l; i++ {
		current := history(c.c.At(i))
		switch mode {
		case modeDefault:
			switch current {
			case bModePath:
				mode = modePath
			case bModeHelpers:
				mode = modeHelper
			case bSwipeUp:
				onSwipe(SwipeDirectionUp, j)
				j++
			case bSwipeRight:
				onSwipe(SwipeDirectionRight, j)
				j++
			case bSwipeDown:
				onSwipe(SwipeDirectionDown, j)
				j++
			case bSwipeLeft:
				onSwipe(SwipeDirectionLeft, j)
				j++
			default:
				return fmt.Errorf("Failed to map in mode, got unexpected value at triplet-index %d, %d %s", i, current, name(mode, current))
			}
		case modePath:
			if len(path) == 0 {
				switch c.tripletsUsedForPathIndex {
				case 1:
					t := int(c.c.At(i))
					path = append(path, t)
				case 2:
					t := int(c.c.At(i)<<3 | c.c.At(i+1))
					i++
					path = append(path, t)
				case 3:
					t := int(c.c.At(i)<<6 | c.c.At(i+1)<<3 | c.c.At(i+2))
					i += 2
					path = append(path, t)
				default:
					panic(fmt.Sprintf("NotImplemeted:CompactHistory:Iterate:modePath tripletsUsedForPathIndex=%d (%dx%d)", c.tripletsUsedForPathIndex, c.gameColumns, c.gameRows))
				}
				continue
			}
			switch current {
			case bitgroupModePathUp, bitgroupModePathUpEnd:
				path = append(path, path[len(path)-1]-c.gameColumns)
			case bitgroupModePathRight, bitgroupModePathRightEnd:
				path = append(path, path[len(path)-1]+1)
			case bitgroupModePathDown, bitgroupModePathDownEnd:
				path = append(path, path[len(path)-1]+c.gameColumns)
			case bitgroupModePathLeft, bitgroupModePathLeftEnd:
				path = append(path, path[len(path)-1]-1)
			default:
				return fmt.Errorf("Failed to map in mode, got unexpected value at triplet-index %d, %d %s", i, current, name(mode, current))
			}
			switch current {
			case bitgroupModePathUpEnd, bitgroupModePathRightEnd, bitgroupModePathDownEnd, bitgroupModePathLeftEnd:
				onCombinePath(path, j)
				path = []int{}
				j++
				mode = modeDefault
			}
		case modeHelper:
			switch current {
			case bitgroupModeHelperHint:
				onHelper(helperHint, j)
				j++
			case bitgroupModeHelperUndo:
				onHelper(helperUndo, j)
				j++
			case bitgroupModeHelperSwap:
				onHelper(helperSwap, j)
				j++

			default:
				return fmt.Errorf("Failed to map in mode '%s', got unexpected value at triplet-index %d, %d %s", mode, i, current, name(mode, current))
			}
			mode = modeDefault
		}
	}
	return nil
}

// Packs a path of indexes.
// Expects that the path is already validated, so that each subsequent element is a "neighbour" of the previous
func combinePathRelative(a, b int) history {
	// b is to the right of a
	if b == a+1 {
		return bitgroupModePathRight
	}
	// b is to the left of a
	if b == a-1 {
		return bitgroupModePathLeft
	}
	// b is on top of a
	if b < a {
		return bitgroupModePathUp
	}
	// b is below a
	return bitgroupModePathDown
}
func relativePathToAbsolute(columns int, previousPosition int, direction history) (int, error) {
	switch direction {
	case bitgroupModePathUp:
		return previousPosition - columns, nil
	case bitgroupModePathRight:
		return previousPosition + 1, nil
	case bitgroupModePathDown:
		return previousPosition + columns, nil
	case bitgroupModePathLeft:
		return previousPosition + columns, nil
	}

	return 0, fmt.Errorf("mapping failure")

}
