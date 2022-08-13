package tallylogic

import (
	"errors"
	"fmt"
	"sort"
)

func copyAndSortSliceToString(slice []int) string {
	var hi []int
	hi = append(hi, slice...)
	sort.Ints(hi)
	return fmt.Sprintf("%v", hi)
}

func remove(slice []int, s int) []int {
	return append(slice[:s], slice[s+1:]...)
}

func (g *hintCalculator) GetHints() (hints []hint) {
	valueForIndexMap := map[int]int64{}
	for i, v := range g.Cells() {
		valueForIndexMap[i] = v.Value()
	}
	for i := 0; i < len(valueForIndexMap); i++ {
		hints = append(hints, g.getHints(valueForIndexMap, []int{i})...)
	}

outer:
	for i := len(hints) - 1; i >= 0; i-- {

		hintI := copyAndSortSliceToString(hints[i].Path)
		for j := 0; j < len(hints); j++ {
			if i == j {
				continue
			}
			if hintI == copyAndSortSliceToString(hints[j].Path) {
				hints = append(hints[:i], hints[i+1:]...)
				continue outer
			}

		}

	}
	// TODO: remove duplicates (where path is just reversed)
	return hints
}

type hint struct {
	Value  int64
	Method EvalMethod
	Path   []int
}

type CellRetriever interface {
	Cells() []Cell
}
type NeighbourRetriever interface {
	NeighboursForCellIndex(index int) ([]int, bool)
}
type Evaluator interface {
	EvaluatesTo(indexes []int, commit bool) (int64, EvalMethod, error)
}

type hintCalculator struct {
	CellRetriever
	NeighbourRetriever
	Evaluator
}

func NewHintCalculator(c CellRetriever, n NeighbourRetriever, e Evaluator) hintCalculator {
	return hintCalculator{c, n, e}
}

func (g *hintCalculator) getHints(valueForIndexMap map[int]int64, path []int) []hint {
	var hints []hint
	neightbours, ok := g.NeighboursForCellIndex(path[len(path)-1])
	if !ok {
		return hints
	}
outer:
	for _, neightbourIndex := range neightbours {
		// Remove already visited
		for _, p := range path {
			if p == neightbourIndex {
				continue outer
			}
		}
		var newPath = path
		newPath = append(newPath, neightbourIndex)
		value, method, err := g.EvaluatesTo(newPath, false)
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
			hints = append(hints, hint{
				Value:  value * 2,
				Method: method,
				Path:   newPath,
			})
		}
		moreHints := g.getHints(valueForIndexMap, newPath)
		if len(moreHints) > 0 {
			hints = append(hints, moreHints...)
		}
	}

	return hints
}
