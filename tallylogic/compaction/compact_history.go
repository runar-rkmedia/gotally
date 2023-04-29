package tallylogiccompaction

import (
	"fmt"

	"github.com/runar-rkmedia/gotally/tallylogic"
)

type history = byte

const (
	bModePath history = iota
	bModeHelpers
	bSwipeUp
	bSwipeRight
	bSwipeDown
	bSwipeLeft
)

const (
	bitgroupModePathToDefault history = iota
	bitgroupModePathToHelpers
	bitgroupModePathUp
	bitgroupModePathRight
	bitgroupModePathDown
	bitgroupModePathLeft
)
const (
	bitgroupModeHelperHint history = iota
	bitgroupModeHelperUndo
	bitgroupModeHelperSwap
)

type CompactHistory struct {
	c CompactTriplets
}

func NewCompactHistory() CompactHistory {
	return CompactHistory{[]byte{}}
}

func (c *CompactHistory) AddHint() {
	c.c.Append(bModeHelpers)
	c.c.Append(bitgroupModeHelperHint)
}
func (c *CompactHistory) AddSwipe(dir tallylogic.SwipeDirection) {
	switch dir {
	case tallylogic.SwipeDirectionUp:
		c.c.Append(bSwipeUp)
	case tallylogic.SwipeDirectionRight:
		c.c.Append(bSwipeRight)
	case tallylogic.SwipeDirectionDown:
		c.c.Append(bSwipeDown)
	case tallylogic.SwipeDirectionLeft:
		c.c.Append(bSwipeLeft)
	}
}
func (c *CompactHistory) AddPath(path []int) error {
	if len(path) < 2 {
		return fmt.Errorf("Path must of at least of length 2")
	}
	c.c.Append(bModePath)
	first := byte(path[0])
	a := first << 2 >> 5
	b := first << 5 >> 5
	c.c.Append(a)
	c.c.Append(b)
	for i := 1; i < len(path); i++ {
		// TODO: turn into direction from previous

	}

	c.c.Append(bitgroupModePathToHelpers)

	return nil
}
