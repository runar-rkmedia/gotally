package tallylogic

import (
	"errors"
	"strconv"
	"strings"
)

func (g *hintCalculator) GetHints() map[string]Hint {
	cells := g.Cells()
	length := len(cells)
	valueForIndex := make([]int64, length)
	neighboursForIndex := make([][]int, length)
	hints := map[string]Hint{}
	for i := 0; i < length; i++ {
		valueForIndex[i] = cells[i].Value()
		n, _ := g.NeighboursForCellIndex(i)
		neighboursForIndex[i] = n
	}
	ch := make(chan Hint)
	doneCh := make(chan struct{}, length)
	done := 0
	for i := 0; i < length; i++ {
		go func(i int) {
			g.getHints(ch, &valueForIndex, &neighboursForIndex, []int{i})
			doneCh <- struct{}{}
		}(i)
	}
	for {
		select {
		case <-doneCh:
			done++
			if done == length {
				return hints
			}
		case h := <-ch:
			hints[h.pathHash] = h
		}
	}
}

type Hint struct {
	Value    int64
	Method   EvalMethod
	Swipe    SwipeDirection
	Path     []int
	pathHash string
}

type CellRetriever interface {
	Cells() []Cell
}
type NeighbourRetriever interface {
	NeighboursForCellIndex(index int) ([]int, bool)
}
type Evaluator interface {
	EvaluatesTo(indexes []int, commit bool, noValidate bool) (int64, EvalMethod, error)
	SoftEvaluatesTo(indexes []int, targetValue int64) (int64, EvalMethod, error)
}

type hintCalculator struct {
	CellRetriever
	NeighbourRetriever
	Evaluator
}

func NewHintCalculator(c CellRetriever, n NeighbourRetriever, e Evaluator) hintCalculator {
	return hintCalculator{c, n, e}
}

func (g *hintCalculator) getHints(ch chan Hint, valueForIndex *[]int64, neighboursForIndex *[][]int, path []int) {
	neightbours := (*neighboursForIndex)[path[0]]
outer:
	for _, neightbourIndex := range neightbours {
		// Remove empty
		if (*valueForIndex)[neightbourIndex] == 0 {
			continue
		}
		// Remove already visited
		for _, p := range path {
			if p == neightbourIndex {
				continue outer
			}
		}
		var newPath = []int{neightbourIndex}
		newPath = append(newPath, path...)
		value, method, err := g.EvaluatesTo(newPath, false, true)
		if errors.Is(err, ErrResultNoCell) {
			continue
		}
		if errors.Is(err, ErrPathIndexEmptyCell) {
			continue
		}
		if errors.Is(err, ErrResultOvershot) {
			continue
		}
		if value > 0 {
			ch <- NewHint(
				value*2,
				method,
				newPath,
			)
		}
		g.getHints(ch, valueForIndex, neighboursForIndex, newPath)
	}
}

func NewHint(value int64, method EvalMethod, path []int) Hint {
	h := Hint{
		Value:  value,
		Method: method,
		Path:   path,
	}
	h.pathHash = h.Hash()
	return h
}

// Returns a hash used to compare for the path of the hint
// This is mostly used for hashmaps, comparing etc.
func (h Hint) Hash() string {
	if h.pathHash != "" {
		return h.pathHash
	}
	b := strings.Builder{}
	for i := 0; i < len(h.Path); i++ {
		b.WriteString(strconv.FormatInt(int64(h.Path[i]), 36))
		b.WriteString(";")
	}
	return b.String()

}
func (h Hint) AreEqaul(hint Hint) bool {
	return h.pathHash == hint.pathHash
}
