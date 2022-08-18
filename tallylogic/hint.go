package tallylogic

import (
	"errors"
	"sort"
	"strconv"
)

func remove(slice []int, s int) []int {
	return append(slice[:s], slice[s+1:]...)
}

func (g *hintCalculator) GetHints() map[string]Hint {
	valueForIndexMap := map[int]int64{}
	hints := map[string]Hint{}
	for i, v := range g.Cells() {
		valueForIndexMap[i] = v.Value()
	}
	length := len(valueForIndexMap)
	ch := make(chan Hint, length)
	doneCh := make(chan struct{}, length)
	done := 0
	for i := 0; i < len(valueForIndexMap); i++ {
		go func(i int) {
			h := g.getHints(valueForIndexMap, []int{i})
			for _, v := range h {
				ch <- v
			}
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

func (g *hintCalculator) getHints(valueForIndexMap map[int]int64, path []int) []Hint {
	var hints []Hint
	neightbours, ok := g.NeighboursForCellIndex(path[0])
	if !ok {
		return hints
	}
outer:
	for _, neightbourIndex := range neightbours {
		// debug = debug && (neightbourIndex == 22 || neightbourIndex == 21 || neightbourIndex == 20)
		// Remove already visited
		for _, p := range path {
			if p == neightbourIndex {
				continue outer
			}
		}
		var newPath = []int{neightbourIndex}
		newPath = append(newPath, path...)
		// var newPath = path
		// newPath := append(path, neightbourIndex)
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
			hints = append(hints, NewHint(
				value*2,
				method,
				newPath,
			))
		}
		moreHints := g.getHints(valueForIndexMap, newPath)
		if len(moreHints) > 0 {
			hints = append(hints, moreHints...)
		}
	}

	return hints
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

func (h Hint) Hash() string {
	if h.pathHash != "" {
		return h.pathHash
	}
	pathSorted := h.Path
	sort.Ints(pathSorted)
	for _, v := range pathSorted {
		h.pathHash += strconv.FormatInt(int64(v), 36) + ";"
	}
	return h.pathHash
}
func (h Hint) AreEqaul(hint Hint) bool {
	return h.pathHash == hint.pathHash
}
