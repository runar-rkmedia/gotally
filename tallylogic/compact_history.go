package tallylogic

import (
	"errors"
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
	bModePathAlt
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

type Helper string

const (
	helperHint Helper = "Hint"
	helperUndo Helper = "Undo"
	helperSwap Helper = "Swap"
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
	return NewCompactHistoryFromBinary(gameColumns, gameRows, []byte{})
}
func NewCompactHistoryFromBinary(gameColumns, gameRows int, data []byte) CompactHistory {
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
		data,
		gameRows,
		gameColumns,
		bitsUsedForPathIndex,
		tripletsUsedForPathIndex,
	}

}

// Returns the inner bytes (does not copy)
func (c *CompactHistory) MarshalBinary() ([]byte, error) {
	return c.c, nil
}

// Returns the inner bytes (does not copy)
func (c *CompactHistory) Bytes() []byte {
	return c.c
}

// Returns the inner bytes (does not copy)
func (c *CompactHistory) BytesCopy() []byte {
	b := make([]byte, len(c.c))
	copy(b, c.c)
	return b
}
func (c *CompactHistory) Restore(b []byte) error {
	c.c = b
	return nil
}
func (c *CompactHistory) Size() int {
	return c.c.Size()
}
func NewCompactHistoryFromGame(game Game) CompactHistory {
	return NewCompactHistoryFromBinary(game.Rules.SizeX, game.Rules.SizeY, game.History.c)
}

func (c *CompactHistory) IsEmpty() bool {
	return c.c.Length() == 0
}

// Returns the number of instructions.
// Not to be confused with the underlying length of the data-structure.
func (c *CompactHistory) Length() int {
	l := 0
	// TODO: this could perhaps be improved easily
	c.Iterate(
		func(dir SwipeDirection, i int) error { l++; return nil },
		func(path []int, i int) error { l++; return nil },
		func(helper Helper, i int) error { l++; return nil },
	)
	return l
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
func (c *CompactHistory) At(index int) Instruction_ {
	var t Instruction_
	c.IterateKind(func(tag Instruction_, i int) error {
		if i != index {
			return nil
		}
		t = tag
		// Return error just to stop iteration
		return fmt.Errorf("Stop")
	})
	return t
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
	first := byte(path[0])
	lengthReduction := 0
	if c.tripletsUsedForPathIndex == 2 && first < 8 {
		lengthReduction = 1
	}
	// Start with a 0-byte(bModePath),
	// followed by the first index path as triplet-count defined by c.tripletsUsedForPathIndex
	toAppend := make([]byte, length+c.tripletsUsedForPathIndex-lengthReduction)

	switch c.tripletsUsedForPathIndex {
	case 0:
		panic(fmt.Sprintf("Invalid state for CompactHistory: tripletsUsedForPathIndex canot be zero. (Size: %d, Length: %d)", c.c.Size(), c.c.Length()))
	case 1:
		toAppend[1] = first
	case 2:
		if first < 8 {
			toAppend[1] = first
			toAppend[0] = bModePathAlt
		} else {
			toAppend[1] = first & 0b00111000 >> 3
			toAppend[2] = first & 0b00000111
		}
	case 3:
		toAppend[1] = first & 0b11000000 >> 6
		toAppend[2] = first & 0b00111000 >> 3
		toAppend[3] = first & 0b00000111
	default:
		panic(fmt.Sprintf("NotImplemeted:CompactHistory:AddPath tripletsUsedForPathIndex=%d (%d)", c.tripletsUsedForPathIndex, c.bitsUsedForPathIndex))
	}

	for i := 1; i < length; i++ {
		toAppend[i+c.tripletsUsedForPathIndex-lengthReduction] = combinePathRelative(path[i-1], path[i])
		if i == length-1 {
			toAppend[i+c.tripletsUsedForPathIndex-lengthReduction] += 4
		}
	}
	c.c.Append(toAppend...)
	return nil
}

// Describe returns a short, human-readable version of the history.
//
// Each instruction is seperated by ;
//
// Swiping Up:      U
//
// Swiping Right:   R
//
// Swiping Down:    D
//
// Swiping Left:    L
//
// Undo:            Z
//
// Hint:            H
//
// Swap:            S
//
// CombinePath:     Comma-separated indexes
func (c *CompactHistory) Describe() string {
	return c.describe(true)
}
func (c *CompactHistory) DescribeWithoutParams() string {
	return c.describe(false)
}
func (c *CompactHistory) describe(withParams bool) string {
	s := strings.Builder{}
	err := c.Iterate(
		func(dir SwipeDirection, i int) error {
			switch dir {
			case SwipeDirectionUp:
				s.WriteString("U;")
			case SwipeDirectionRight:
				s.WriteString("R;")
			case SwipeDirectionDown:
				s.WriteString("D;")
			case SwipeDirectionLeft:
				s.WriteString("L;")
			default:
				return fmt.Errorf("???")
			}
			return nil
		},
		func(path []int, i int) error {
			l := len(path)
			if !withParams {
				s.WriteString("C;")
				return nil
			}
			for i := 0; i < l; i++ {
				s.WriteString(strconv.FormatInt(int64(path[i]), 10))
				if i < l-1 {
					s.WriteString(",")
				}
			}
			s.WriteString(";")
			return nil
		},
		func(helper Helper, i int) error {
			switch helper {
			case helperHint:
				s.WriteString("H;")
			case helperSwap:
				s.WriteString("S;")
			case helperUndo:
				s.WriteString("Z;")
			default:
				return fmt.Errorf("???")
			}
			return nil
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

type Instruction_ struct {
	IsSwipe   bool
	IsPath    bool
	IsHelper  bool
	Path      []int
	Direction SwipeDirection
	Helper    Helper
}

func NewSwipeInstruction_(direction SwipeDirection) Instruction_ {
	return Instruction_{IsSwipe: true, Direction: direction}
}
func NewPathInstruction_(path []int) Instruction_ {
	return Instruction_{IsPath: true, Path: path}
}
func NewHelperInstruction_(helper Helper) Instruction_ {
	return Instruction_{IsHelper: true, Helper: helper}
}
func (ins Instruction_) IsHelperUndo() bool {
	return ins.IsHelper && ins.Helper == helperUndo
}
func (ins Instruction_) IsHelperHint() bool {
	return ins.IsHelper && ins.Helper == helperHint
}
func (ins Instruction_) String() string {
	switch {
	case ins.IsSwipe:
		return string(ins.Direction)
	case ins.IsHelper:
		return string(ins.Helper)
	case ins.IsPath:
		s := ""
		for i := 0; i < len(ins.Path); i++ {
			s += strconv.FormatInt(int64(ins.Path[i]), 10) + ","
		}
		return s[:len(s)-1]
	}
	return "Instruction???"
}
func (ins Instruction_) Equal(b Instruction_) bool {
	switch {
	case ins.IsSwipe:
		return ins.IsSwipe && ins.Direction == b.Direction
	case ins.IsHelper:
		return ins.IsHelper && ins.Helper == b.Helper
	case ins.IsPath:
		if !b.IsSwipe {
			return false
		}
		if len(ins.Path) != len(b.Path) {
			return false
		}
		for i := 0; i < len(ins.Path); i++ {
			if ins.Path[i] != b.Path[i] {
				return false
			}
		}
		return true
	}
	return false
}

func (c *CompactHistory) All() ([]Instruction_, error) {
	instructions := []Instruction_{}
	err := c.IterateKind(func(tag Instruction_, i int) error {
		instructions = append(instructions, tag)
		return nil
	})

	return instructions, err
}
func (c *CompactHistory) Last() (Instruction_, error) {
	// naive implementation, could easily be improved
	all, err := c.All()
	if err != nil {
		return Instruction_{}, fmt.Errorf("failed to retrieve the last item: %v", err)
	}
	return all[len(all)-1], nil
}
func (c *CompactHistory) IterateKind(
	f func(tag Instruction_, i int) error,
) error {
	return c.Iterate(
		func(dir SwipeDirection, i int) error {
			return f(Instruction_{Direction: dir, IsSwipe: true}, i)
		},
		func(path []int, i int) error {
			return f(Instruction_{Path: path, IsPath: true}, i)
		},
		func(helper Helper, i int) error {
			return f(Instruction_{Helper: helper, IsHelper: true}, i)
		},
	)
}

var (
	ErrNoMoreHistoryToUndo = errors.New("No more history to undo")
)

// helper for creating a copy of the history where the undo-actions and its targets are removed
// For instance, Say we have the following History:
// Input: L;R;Z;
// Output: L;
// Input: D;L;R;Z;U;Z;Z;
// Output: D;
func (c CompactHistory) FilterForUndo(appendUndo bool) ([]Instruction_, error) {
	history, err := c.All()
	if err != nil {
		return []Instruction_{}, fmt.Errorf("failed to undo game: %w", err)
	}
	l := len(history)
	undoes := 0
	others := 0
	for i := 0; i < len(history); i++ {
		if history[i].IsHelperUndo() {
			undoes++
		} else {
			others++
		}
	}
	if undoes >= others {
		return history, ErrNoMoreHistoryToUndo
	}
	if appendUndo {
		history = append(history, NewHelperInstruction_(helperUndo))
	}
	var dropped int
	for i := 0; i < len(history); i++ {
		if history[i].IsHelperUndo() {
			if dropped == l {
				return []Instruction_{}, nil
			}
			if i >= i {
				history = append(history[:i-1], history[i+1:]...)
				dropped++
			} else {
				history = append(history[:i], history[i+1:]...)
				panic("what")
			}
			i = -1
		}
	}
	return history, nil
}
func (c *CompactHistory) Iterate(
	onSwipe func(dir SwipeDirection, i int) error,
	onCombinePath func(path []int, i int) error,
	onHelper func(helper Helper, i int) error,
) error {
	l := c.c.Length()
	mode := modeDefault
	alt := false
	var j int
	path := []int{}
	for i := 0; i < l; i++ {
		current := history(c.c.TripletAt(i))
		switch mode {
		case modeDefault:
			switch current {
			case bModePath:
				mode = modePath
				alt = false
			case bModePathAlt:
				mode = modePath
				alt = true
			case bModeHelpers:
				mode = modeHelper
			case bSwipeUp:
				err := onSwipe(SwipeDirectionUp, j)
				if err != nil {
					return err
				}
				j++
			case bSwipeRight:
				err := onSwipe(SwipeDirectionRight, j)
				if err != nil {
					return err
				}
				j++
			case bSwipeDown:
				err := onSwipe(SwipeDirectionDown, j)
				if err != nil {
					return err
				}
				j++
			case bSwipeLeft:
				err := onSwipe(SwipeDirectionLeft, j)
				if err != nil {
					return err
				}
				j++
			default:
				return fmt.Errorf("Failed to map in mode, got unexpected value at triplet-index %d, %d %s", i, current, name(mode, current))
			}
		case modePath:
			if len(path) == 0 {
				switch c.tripletsUsedForPathIndex {
				case 1:
					t := int(c.c.TripletAt(i))
					path = append(path, t)
				case 2:
					if alt {
						t := int(c.c.TripletAt(i))
						path = append(path, t)

					} else {

						t := int(c.c.TripletAt(i)<<3 | c.c.TripletAt(i+1))
						i++
						path = append(path, t)
					}
				case 3:
					t := int(c.c.TripletAt(i)<<6 | c.c.TripletAt(i+1)<<3 | c.c.TripletAt(i+2))
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
				err := onCombinePath(path, j)
				if err != nil {
					return err
				}
				path = []int{}
				j++
				mode = modeDefault
			}
		case modeHelper:
			switch current {
			case bitgroupModeHelperHint:
				err := onHelper(helperHint, j)
				if err != nil {
					return err
				}
				j++
			case bitgroupModeHelperUndo:
				err := onHelper(helperUndo, j)
				if err != nil {
					return err
				}
				j++
			case bitgroupModeHelperSwap:
				err := onHelper(helperSwap, j)
				if err != nil {
					return err
				}
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
