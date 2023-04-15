package tallylogic

import (
	"errors"
	"fmt"
	"strings"

	"github.com/runar-rkmedia/gotally/tallylogic/cell"
	"golang.org/x/net/context"
)

func (g *hintCalculator) GetHint() *Hint {
	for _, v := range g.GetNHints(1) {
		return &v
	}
	return nil
}
func (g *hintCalculator) GetHints() map[string]Hint {
	return g.GetNHints(0)
}

func (g *hintCalculator) GetNHints(maxHints int) map[string]Hint {
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
	ctx, cancel := context.WithCancel(context.TODO())
	for i := 0; i < length; i++ {
		go func(i int) {
			g.getHints(ctx, ch, &valueForIndex, &neighboursForIndex, []int{i})
			doneCh <- struct{}{}
		}(i)
	}
	for {
		select {
		case <-doneCh:
			done++
			if done == length {
				cancel()
				return hints
			}
		case h := <-ch:
			hints[h.pathHash] = h
			if maxHints > 0 && maxHints >= len(hints) {
				cancel()
				return hints
			}
		}
	}
}

// Retrieves hints in a consistant order
func (g *hintCalculator) GetNHintsConsistant(ctx context.Context, maxHints int) []Hint {
	cells := g.Cells()
	maxHints = 1
	length := len(cells)
	valueForIndex := make([]int64, length)
	neighboursForIndex := make([][]int, length)
	hints := []Hint{}
	for i := 0; i < length; i++ {
		valueForIndex[i] = cells[i].Value()
		n, _ := g.NeighboursForCellIndex(i)
		neighboursForIndex[i] = n
	}
	ch := make(chan Hint)
	doneCh := make(chan struct{}, length)
	done := 0
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		for i := 0; i < length; i++ {
			g.getHints(ctx, ch, &valueForIndex, &neighboursForIndex, []int{i})
			doneCh <- struct{}{}
		}
	}()
	for {
		select {
		case <-doneCh:
			done++
			if done == length {
				cancel()
				return hints
			}
		case h := <-ch:
			hints = append(hints, h)
			if maxHints > 0 && maxHints >= len(hints) {
				cancel()
				return hints
			}
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
	Cells() []cell.Cell
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

func (g *hintCalculator) getHints(ctx context.Context, ch chan Hint, valueForIndex *[]int64, neighboursForIndex *[][]int, path []int) {
	neightbours := (*neighboursForIndex)[path[0]]
outer:
	for _, neightbourIndex := range neightbours {
		select {
		case <-ctx.Done():
			return
		default:
		}
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
		g.getHints(ctx, ch, valueForIndex, neighboursForIndex, newPath)
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
		b.WriteByte(byte(h.Path[i]))
	}
	return b.String()
}
func (h Hint) String() string {
	if h.Swipe != "" {
		return "Swipe: " + string(h.Swipe)
	}

	method := "multiplication"
	if h.Method == EvalMethodSum {
		method = "addition"
	}
	return fmt.Sprintf("Combine path by %s for %d (%v)", method, h.Value, h.Path)
}
func (h Hint) AreEqaul(hint Hint) bool {
	return h.pathHash == hint.pathHash
}
